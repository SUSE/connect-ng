package connect

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAnnounceSystem(t *testing.T) {
	createTestCredentials("", "", t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("System-Token", "token")
		io.WriteString(w, `{"login":"test-user","password":"test-password"}`)
	}))
	defer ts.Close()

	CFG.BaseURL = ts.URL
	user, password, err := announceSystem(nil)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if user != "test-user" {
		t.Errorf("Expected user: \"test-user\", got: \"%s\"", user)
	}
	if password != "test-password" {
		t.Errorf("Expected password: \"test-password\", got: \"%s\"", password)
	}

	// System token should have been updated.
	creds, err := getCredentials()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if creds.SystemToken != "token" {
		t.Fatalf("Unexpected token '%v', should have been 'token'", creds.SystemToken)
	}
}

func TestGetActivations(t *testing.T) {
	response := readTestFile("activations.json", t)
	createTestCredentials("", "", t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}))
	defer ts.Close()

	CFG.BaseURL = ts.URL
	activations, err := systemActivations()
	if err != nil {
		t.Errorf("%s", err)
	}
	if l := len(activations); l != 1 {
		t.Errorf("len(activations) == %d, expected 1", l)
	}
	key := "SUSE-MicroOS/5.0/x86_64"
	if _, ok := activations[key]; !ok {
		t.Errorf("activations map missing key [%s]", key)
	}
}

func TestGetActivationsRequest(t *testing.T) {
	var (
		user       = "testuser"
		password   = "testpassword"
		url        = "/connect/systems/activations"
		gotRequest *http.Request
	)
	createTestCredentials(user, password, t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequest = r // make request available outside this func after call
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, "[]")
	}))
	defer ts.Close()

	CFG.BaseURL = ts.URL
	if _, err := systemActivations(); err != nil {
		t.Fatalf("Unexpected error [%s]", err)
	}

	gotURL := gotRequest.URL.String()
	u, p, ok := gotRequest.BasicAuth()
	if !ok || u != user || p != password || gotURL != url {
		t.Errorf("Server got request with %s, %s, %s. Expected %s, %s, %s", u, p, gotURL, user, password, url)
	}
}

func TestGetActivationsError(t *testing.T) {
	createTestCredentials("", "", t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusInternalServerError)
	}))
	defer ts.Close()

	CFG.BaseURL = ts.URL
	if _, err := systemActivations(); err == nil {
		t.Error("Expecting error. Got none.")
	}
}

func TestUpToDateOkay(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusUnprocessableEntity)
	}))
	defer ts.Close()

	CFG.BaseURL = ts.URL
	if !UpToDate() {
		t.Error("Expecting UpToDate()==true, got false")
	}
}

func TestGetProduct(t *testing.T) {
	createTestCredentials("", "", t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(readTestFile("product.json", t))
	}))
	defer ts.Close()

	CFG.BaseURL = ts.URL
	productQuery := Product{Name: "SLES", Version: "15.2", Arch: "x86_64"}
	product, err := showProduct(productQuery)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if l := len(product.Extensions); l != 1 {
		t.Fatalf("len(product.Extensions) == %d, expected 1", l)
	}
	if l := len(product.Extensions[0].Extensions); l != 8 {
		t.Fatalf("len(product.Extensions[0].Extensions) == %d, expected 8", l)
	}

}

func TestGetProductError(t *testing.T) {
	createTestCredentials("", "", t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errMsg := "{\"status\":422,\"error\":\"No product specified\",\"type\":\"error\",\"localized_error\":\"No product specified\"}"
		http.Error(w, errMsg, http.StatusUnprocessableEntity)
	}))
	defer ts.Close()

	CFG.BaseURL = ts.URL
	productQuery := Product{Name: "Dummy"}
	_, err := showProduct(productQuery)
	if ae, ok := err.(APIError); ok {
		if ae.Code != http.StatusUnprocessableEntity {
			t.Fatalf("Expecting APIError(422). Got %s", err)
		}
	}
}

func TestUpgradeProduct(t *testing.T) {
	createTestCredentials("", "", t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(readTestFile("service.json", t))
	}))
	defer ts.Close()

	CFG.BaseURL = ts.URL
	product := Product{Name: "SUSE-MicroOS", Version: "5.0", Arch: "x86_64"}
	service, err := upgradeProduct(product)
	if err != nil {
		t.Fatalf("%s", err)
	}
	name := "SUSE_Linux_Enterprise_Micro_5.0_x86_64"
	if service.Name != name {
		t.Fatalf("Expecting service name %s. Got %s", name, service.Name)
	}
	if service.Product.ToTriplet() != product.ToTriplet() {
		t.Fatalf("Unexpected product %s", service.Product.ToTriplet())
	}
}

func TestUpgradeProductError(t *testing.T) {
	createTestCredentials("", "", t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errMsg := "{\"status\":422,\"error\":\"No product specified\",\"type\":\"error\",\"localized_error\":\"No product specified\"}"
		http.Error(w, errMsg, http.StatusUnprocessableEntity)
	}))
	defer ts.Close()

	CFG.BaseURL = ts.URL
	product := Product{Name: "Dummy"}
	_, err := upgradeProduct(product)
	if ae, ok := err.(APIError); ok {
		if ae.Code != http.StatusUnprocessableEntity {
			t.Fatalf("Expecting APIError(422). Got %s", err)
		}
	}
}

func TestDeactivateProduct(t *testing.T) {
	createTestCredentials("", "", t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(readTestFile("service_inactive.json", t))
	}))
	defer ts.Close()

	CFG.BaseURL = ts.URL
	product := Product{Name: "sle-module-basesystem", Version: "15.2", Arch: "x86_64"}
	service, err := deactivateProduct(product)
	if err != nil {
		t.Fatalf("%s", err)
	}
	name := "Basesystem_Module_15_SP2_x86_64"
	if service.Name != name {
		t.Fatalf("Expecting service name %s. Got %s", name, service.Name)
	}
	if service.Product.ToTriplet() != product.ToTriplet() {
		t.Fatalf("Unexpected product %s", service.Product.ToTriplet())
	}
}

func TestDeactivateProductSMT(t *testing.T) {
	createTestCredentials("", "", t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(readTestFile("service_inactive_smt.json", t))
	}))
	defer ts.Close()

	CFG.BaseURL = ts.URL
	product := Product{Name: "SUSE-MicroOS", Version: "5.0", Arch: "x86_64"}
	service, err := deactivateProduct(product)
	if err != nil {
		t.Fatalf("%s", err)
	}
	name := "SMT_DUMMY_NOREMOVE_SERVICE"
	if service.Name != name {
		t.Fatalf("Expecting service name %s. Got %s", name, service.Name)
	}
	if !service.Product.isEmpty() {
		t.Fatalf("Unexpected product %s", service.Product.ToTriplet())
	}
}

func TestDeactivateProductError(t *testing.T) {
	createTestCredentials("", "", t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errMsg := "{\"status\":422,\"error\":\"No product specified\",\"type\":\"error\",\"localized_error\":\"No product specified\"}"
		http.Error(w, errMsg, http.StatusUnprocessableEntity)
	}))
	defer ts.Close()

	CFG.BaseURL = ts.URL
	product := Product{Name: "Dummy"}
	_, err := deactivateProduct(product)
	if ae, ok := err.(APIError); ok {
		if ae.Code != http.StatusUnprocessableEntity {
			t.Fatalf("Expecting APIError(422). Got %s", err)
		}
	}
}

func TestProductMigrations(t *testing.T) {
	createTestCredentials("", "", t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(readTestFile("migrations.json", t))
	}))
	defer ts.Close()
	CFG.BaseURL = ts.URL

	migrations, err := productMigrations(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(migrations) != 2 {
		t.Fatalf("len(migrations) == %d, expected 2", len(migrations))
	}
}

func TestProductMigrationsSMT(t *testing.T) {
	createTestCredentials("", "", t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(readTestFile("migrations-smt.json", t))
	}))
	defer ts.Close()
	CFG.BaseURL = ts.URL

	migrations, err := productMigrations(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(migrations) != 1 {
		t.Fatalf("len(migrations) == %d, expected 1", len(migrations))
	}
	gotID := migrations[0][0].ID
	expectedID := 101361
	if gotID != expectedID {
		t.Fatalf("Got ID: %d, expected: %d", gotID, expectedID)
	}
}

func createTestUptimeLogFileWithContent(content string) (string, error) {
	tempFile, err := ioutil.TempFile("", "testUptimeLog")
	if err != nil {
		return "", err
	}
	tempFilePath := tempFile.Name()
	if _, err := tempFile.WriteString(content); err != nil {
		os.Remove(tempFilePath)
		return "", err
	}
	if err := tempFile.Close(); err != nil {
		os.Remove(tempFilePath)
		return "", err
	}

	return tempFilePath, nil
}

func TestUptimeLogFileDoesNotExist(t *testing.T) {
	uptimeLogFileContent := ``
	tempFilePath, err := createTestUptimeLogFileWithContent(uptimeLogFileContent)
	if err != nil {
		t.Fatalf("Failed to create temp uptime log file for testing")
	}
	os.Remove(tempFilePath)
	uptimeLog, err := readUptimeLogFile(tempFilePath)
	if uptimeLog != nil && err != nil {
		t.Fatalf("Expected uptime log and err to be nil if uptime log file does not exist")
	}
}

func TestReadUptimeLogFile(t *testing.T) {
	uptimeLogFileContent := `2024-01-18:000000000000001000110000
2024-01-13:000000000000000000010000`
	tempFilePath, err := createTestUptimeLogFileWithContent(uptimeLogFileContent)
	if err != nil {
		t.Fatalf("Failed to create temp uptime log file for testing")
	}
	defer os.Remove(tempFilePath)
	uptimeLog, err := readUptimeLogFile(tempFilePath)
	if err != nil {
		t.Fatalf("Failed to read uptime log file")
	}
	if uptimeLog == nil {
		t.Fatal("Failed to open uptime log file")
	}
	if len(uptimeLog) != 2 {
		t.Fatalf("Expected two entries in uptime log, got %#v instead", len(uptimeLog))
	}
}
