package connect

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetActivations(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		f, err := os.Open("../testdata/activations.json")
		if err != nil {
			t.Fatal(err)
		}
		io.Copy(w, f)
	}))
	defer ts.Close()

	config := Config{BaseURL: ts.URL}
	creds := Credentials{}
	activations, err := GetActivations(config, creds)
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

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequest = r // make request available outside this func after call
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, "[]")
	}))
	defer ts.Close()

	config := Config{BaseURL: ts.URL}
	creds := Credentials{Username: user, Password: password}
	if _, err := GetActivations(config, creds); err != nil {
		t.Errorf("Unexpected error [%s]", err)
	}

	gotURL := gotRequest.URL.String()
	u, p, ok := gotRequest.BasicAuth()
	if !ok || u != user || p != password || gotURL != url {
		t.Errorf("Server got request with %s, %s, %s. Expected %s, %s, %s", u, p, gotURL, user, password, url)
	}
}

func TestGetActivationsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusInternalServerError)
	}))
	defer ts.Close()

	config := Config{BaseURL: ts.URL}
	creds := Credentials{}
	if _, err := GetActivations(config, creds); err == nil {
		t.Error("Expecting error. Got none.")
	}
}
