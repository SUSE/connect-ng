package connect

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoGetInsecure(t *testing.T) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.StartTLS()
	defer ts.Close()

	config := Config{BaseURL: ts.URL, Insecure: false}
	_, err := DoGET(config, Credentials{}, "/")
	if err == nil {
		t.Error("Expecting certificate error. Got none.")
	}

	config.Insecure = true
	_, err = DoGET(config, Credentials{}, "/")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}
