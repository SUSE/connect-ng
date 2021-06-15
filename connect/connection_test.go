package connect

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoGetInsecure(t *testing.T) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.StartTLS()
	defer ts.Close()

	CFG.BaseURL = ts.URL
	CFG.Insecure = false
	_, err := DoGET(Credentials{}, "/")
	if err == nil {
		t.Error("Expecting certificate error. Got none.")
	}

	CFG.Insecure = true
	_, err = DoGET(Credentials{}, "/")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestParseError(t *testing.T) {
	s := `{"type":"error","error":"Invalid system credentials","localized_error":"Invalid system credentials"}`
	body := strings.NewReader(s)
	expected := "Invalid system credentials"
	got := parseError(body)
	if got != expected {
		t.Errorf("parseError(), got: %s, expected: %s", got, expected)
	}
}

func TestParseErrorNotJson(t *testing.T) {
	body := strings.NewReader("not json")
	got := parseError(body)
	if got != "" {
		t.Errorf("parseError(), got: %s, expected \"\"", got)
	}
}
