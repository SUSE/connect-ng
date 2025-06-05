package connect

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/SUSE/connect-ng/internal/collectors"
	"github.com/SUSE/connect-ng/internal/credentials"
	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/assert"
)

// TODO(mssola): remove before finishing RR4.
func shittyGlobalVariableNeededForNow() {
	CFG = DefaultOptions()
}

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

func TestGetActivationsRequest(t *testing.T) {
	shittyGlobalVariableNeededForNow()

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
	shittyGlobalVariableNeededForNow()

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

func TestProductMigrations(t *testing.T) {
	shittyGlobalVariableNeededForNow()

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
	shittyGlobalVariableNeededForNow()

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
	shittyGlobalVariableNeededForNow()

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

func mockDetectArchitecture() {
	collectors.DetectArchitecture = func() (string, error) {
		return "x86_64", nil
	}
}

func TestMakeSysInfoBody(t *testing.T) {
	shittyGlobalVariableNeededForNow()

	assert := assert.New(t)
	expectedBody := strings.TrimSpace(string(util.ReadTestFile("api/system_information_body.json", t)))

	// lets overwrite the collectors here to not really call any detection in unit tests
	mandatoryCollectors = []collectors.Collector{
		collectors.FakeCollectorNew("hostname", "localhost"),
		collectors.FakeCollectorNew("cpus", 2),
		collectors.FakeCollectorNew("sockets", 2),
		collectors.FakeCollectorNew("mem_total", 1337),
	}
	mockDetectArchitecture()

	body, err := makeSysInfoBody("sle-15-x86_64", "test", []byte{}, false)

	assert.NoError(err)
	assert.Equal(expectedBody, string(body))
}
