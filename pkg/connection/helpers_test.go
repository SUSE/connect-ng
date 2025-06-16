package connection

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func fixture(t *testing.T, path string) []byte {
	t.Helper()

	absolut, pathErr := filepath.Abs(filepath.Join("../../testdata", path))
	if pathErr != nil {
		t.Fatalf("Could not build fixture path from %s", path)
	}

	data, err := os.ReadFile(absolut)
	if err != nil {
		t.Fatalf("Could not read fixture: %s", err)
	}
	return data

}

func NewTestServerSetupWith(t *testing.T, verb, path string, setup func(http.ResponseWriter)) *httptest.Server {
	assert := assert.New(t)

	handler := func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(verb, request.Method, "HTTP method matches")
		assert.Equal(path, request.URL.Path, "URL path matches")

		setup(response)
	}
	return httptest.NewServer(http.HandlerFunc(handler))
}

func NewTestTLSServerSetupWith(t *testing.T, verb, path string, setup func(http.ResponseWriter)) *httptest.Server {
	assert := assert.New(t)

	handler := func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(verb, request.Method, "HTTP method matches")
		assert.Equal(path, request.URL.Path, "URL path matches")

		setup(response)
	}

	return httptest.NewTLSServer(http.HandlerFunc(handler))
}
