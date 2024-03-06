package connect

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/SUSE/connect-ng/internal/credentials"
	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/assert"
)

// NOTE: Until there is a better implementation of the credentials package
//
//	we need to set the file creation path for SCCCredentials to /tmp
//	to allow creating these files in this test.
//	This is not nice but creating stubs with this current implemented
//	API is almost impossible since you need mock the whole object, resulting
//	in a complete rewrite.
func setRootToTmp() {
	CFG.FsRoot = "/tmp"
}

func TestAnnounceSystem(t *testing.T) {
	assert := assert.New(t)

	setRootToTmp()
	credentials.CreateTestCredentials("", "", CFG.FsRoot, t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("System-Token", "token")
		io.WriteString(w, `{"login":"test-user","password":"test-password"}`)
	}))
	defer ts.Close()
	CFG.BaseURL = ts.URL

	user, password, err := announceSystem(nil)
	assert.NoError(err)
	assert.Equal("test-user", user)
	assert.Equal("test-password", password)

	// System token should have been updated.
	creds, err := credentials.ReadCredentials(credentials.SystemCredentialsPath(CFG.FsRoot))
	assert.NoError(err)
	assert.Equal("token", creds.SystemToken, "system token mismatch")
}

func TestGetActivations(t *testing.T) {
	assert := assert.New(t)
	response := util.ReadTestFile("activations.json", t)

	setRootToTmp()
	credentials.CreateTestCredentials("", "", CFG.FsRoot, t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}))
	defer ts.Close()
	CFG.BaseURL = ts.URL

	activations, err := systemActivations()
	assert.NoError(err)
	assert.Len(activations, 1, "expected 1 activation")
	assert.Contains(activations, "SUSE-MicroOS/5.0/x86_64")
}

func TestGetActivationsRequest(t *testing.T) {
	var gotRequest *http.Request

	assert := assert.New(t)
	expectedUser := "test-user"
	expectedPassword := "test-password"
	expectedURL := "/connect/systems/activations"

	setRootToTmp()
	credentials.CreateTestCredentials(expectedUser, expectedPassword, CFG.FsRoot, t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequest = r // make request available outside this func after call
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, "[]")
	}))
	defer ts.Close()
	CFG.BaseURL = ts.URL

	_, err := systemActivations()
	assert.NoError(err)

	actualURL := gotRequest.URL.String()
	user, password, ok := gotRequest.BasicAuth()

	assert.True(ok)
	assert.Equal(expectedUser, user)
	assert.Equal(expectedPassword, password)
	assert.Equal(expectedURL, actualURL)
}

func TestGetActivationsError(t *testing.T) {
	assert := assert.New(t)

	setRootToTmp()
	credentials.CreateTestCredentials("", "", CFG.FsRoot, t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusInternalServerError)
	}))
	defer ts.Close()
	CFG.BaseURL = ts.URL

	_, err := systemActivations()
	assert.ErrorContains(err, "(500)")
}

func TestUpToDateOkay(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusUnprocessableEntity)
	}))
	defer ts.Close()
	CFG.BaseURL = ts.URL

	assert.True(UpToDate())
}

func TestGetProduct(t *testing.T) {
	assert := assert.New(t)
	productQuery := Product{Name: "SLES", Version: "15.2", Arch: "x86_64"}

	setRootToTmp()
	credentials.CreateTestCredentials("", "", CFG.FsRoot, t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(util.ReadTestFile("product.json", t))
	}))
	defer ts.Close()
	CFG.BaseURL = ts.URL

	product, err := showProduct(productQuery)
	assert.NoError(err)
	assert.Len(product.Extensions, 1)
	assert.Len(product.Extensions[0].Extensions, 8)
}

func TestGetProductError(t *testing.T) {
	assert := assert.New(t)
	productQuery := Product{Name: "Dummy"}

	setRootToTmp()
	credentials.CreateTestCredentials("", "", CFG.FsRoot, t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errMsg := "{\"status\":422,\"error\":\"No product specified\",\"type\":\"error\",\"localized_error\":\"No product specified\"}"
		http.Error(w, errMsg, http.StatusUnprocessableEntity)
	}))
	defer ts.Close()
	CFG.BaseURL = ts.URL

	_, err := showProduct(productQuery)
	assert.ErrorContains(err, "(422)", "expected status 422")
}

func TestUpgradeProduct(t *testing.T) {
	assert := assert.New(t)
	product := Product{Name: "SUSE-MicroOS", Version: "5.0", Arch: "x86_64"}
	expectedName := "SUSE_Linux_Enterprise_Micro_5.0_x86_64"
	expectedTriplet := product.ToTriplet()

	setRootToTmp()
	credentials.CreateTestCredentials("", "", CFG.FsRoot, t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(util.ReadTestFile("service.json", t))
	}))
	defer ts.Close()
	CFG.BaseURL = ts.URL

	service, err := upgradeProduct(product)
	assert.NoError(err)
	assert.Equal(expectedName, service.Name)
	assert.Equal(expectedTriplet, service.Product.ToTriplet())
}

func TestUpgradeProductError(t *testing.T) {
	assert := assert.New(t)
	product := Product{Name: "Dummy"}

	setRootToTmp()
	credentials.CreateTestCredentials("", "", CFG.FsRoot, t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errMsg := "{\"status\":422,\"error\":\"No product specified\",\"type\":\"error\",\"localized_error\":\"No product specified\"}"
		http.Error(w, errMsg, http.StatusUnprocessableEntity)
	}))
	defer ts.Close()
	CFG.BaseURL = ts.URL

	_, err := upgradeProduct(product)
	assert.ErrorContains(err, "(422)", "expected status 422")
}

func TestDeactivateProduct(t *testing.T) {
	assert := assert.New(t)
	product := Product{Name: "sle-module-basesystem", Version: "15.2", Arch: "x86_64"}
	expectedName := "Basesystem_Module_15_SP2_x86_64"
	expectedTriplet := product.ToTriplet()

	setRootToTmp()
	credentials.CreateTestCredentials("", "", CFG.FsRoot, t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(util.ReadTestFile("service_inactive.json", t))
	}))
	defer ts.Close()
	CFG.BaseURL = ts.URL

	service, err := deactivateProduct(product)
	assert.NoError(err)
	assert.Equal(expectedName, service.Name)
	assert.Equal(expectedTriplet, service.Product.ToTriplet())
}

func TestDeactivateProductSMT(t *testing.T) {
	assert := assert.New(t)
	product := Product{Name: "SUSE-MicroOS", Version: "5.0", Arch: "x86_64"}
	expectedName := "SMT_DUMMY_NOREMOVE_SERVICE"

	setRootToTmp()
	credentials.CreateTestCredentials("", "", CFG.FsRoot, t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(util.ReadTestFile("service_inactive_smt.json", t))
	}))
	defer ts.Close()
	CFG.BaseURL = ts.URL

	service, err := deactivateProduct(product)
	assert.NoError(err)
	assert.Equal(expectedName, service.Name)
	assert.True(service.Product.isEmpty(), "expected no product")
}

func TestDeactivateProductError(t *testing.T) {
	assert := assert.New(t)
	product := Product{Name: "Dummy"}

	setRootToTmp()
	credentials.CreateTestCredentials("", "", CFG.FsRoot, t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errMsg := "{\"status\":422,\"error\":\"No product specified\",\"type\":\"error\",\"localized_error\":\"No product specified\"}"
		http.Error(w, errMsg, http.StatusUnprocessableEntity)
	}))
	defer ts.Close()
	CFG.BaseURL = ts.URL

	_, err := deactivateProduct(product)
	assert.ErrorContains(err, "(422)", "expected status 422")
}

func TestProductMigrations(t *testing.T) {
	assert := assert.New(t)

	setRootToTmp()
	credentials.CreateTestCredentials("", "", CFG.FsRoot, t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(util.ReadTestFile("migrations.json", t))
	}))
	defer ts.Close()
	CFG.BaseURL = ts.URL

	migrations, err := productMigrations(nil)
	assert.NoError(err)
	assert.Len(migrations, 2, "migrations")
}

func TestProductMigrationsSMT(t *testing.T) {
	assert := assert.New(t)
	expectedID := 101361

	setRootToTmp()
	credentials.CreateTestCredentials("", "", CFG.FsRoot, t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(util.ReadTestFile("migrations-smt.json", t))
	}))
	defer ts.Close()
	CFG.BaseURL = ts.URL

	migrations, err := productMigrations(nil)
	assert.NoError(err)
	assert.Len(migrations, 1, "migrations")
	assert.Equal(expectedID, migrations[0][0].ID)
}

func createTestUptimeLogFileWithContent(content string) (string, error) {
	tempFile, err := os.CreateTemp("", "testUptimeLog")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()
	tempFilePath := tempFile.Name()
	if _, err := tempFile.WriteString(content); err != nil {
		os.Remove(tempFilePath)
		return "", err
	}

	return tempFilePath, nil
}

func TestUptimeLogFileDoesNotExist(t *testing.T) {
	tempFilePath, err := createTestUptimeLogFileWithContent("")
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
		t.Fatalf("Failed to read uptime log file: %s", err)
	}
	if uptimeLog == nil {
		t.Fatal("Failed to open uptime log file")
	}
	if len(uptimeLog) != 2 {
		t.Fatalf("Expected two entries in uptime log, got %#v instead", len(uptimeLog))
	}
}
