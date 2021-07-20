package connect

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCallHTTPSecure(t *testing.T) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.StartTLS()
	defer ts.Close()

	CFG.BaseURL = ts.URL
	CFG.Insecure = false
	_, err := callHTTP("GET", "/", nil, nil, authNone)
	if err == nil {
		t.Error("Expecting certificate error. Got none.")
	}

	httpclient = nil // force new http client+transport creation
	CFG.Insecure = true
	_, err = callHTTP("GET", "/", nil, nil, authNone)
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

func TestParseErrorLocalized(t *testing.T) {
	s := `{"type":"error","error":"No subscription with this Registration Code found",
		  "localized_error":"Keine Subscription mit diesem Registrierungscode gefunden"}`
	body := strings.NewReader(s)
	expected := "Keine Subscription mit diesem Registrierungscode gefunden"
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

func TestAuthToken(t *testing.T) {
	CFG.Token = "test-token"
	var gotRequest *http.Request

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequest = r
	}))
	defer ts.Close()

	CFG.BaseURL = ts.URL
	callHTTP("POST", "", nil, nil, authToken)

	got := gotRequest.Header.Get("Authorization")
	expected := "Token token=test-token"
	if got != expected {
		t.Errorf("Expected: \"%s\", got: \"%s\"", expected, got)
	}
}
