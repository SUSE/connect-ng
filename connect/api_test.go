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
