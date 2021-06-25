package connect

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetActivations(t *testing.T) {
	response := readTestFile("activations.json", t)
	createTestCredentials("", "", t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}))
	defer ts.Close()

	CFG.BaseURL = ts.URL
	activations, err := GetActivations()
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
	if _, err := GetActivations(); err != nil {
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
	if _, err := GetActivations(); err == nil {
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
