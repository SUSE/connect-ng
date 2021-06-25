package connect

import (
	"strings"
	"testing"
)

func TestParseCredentials(t *testing.T) {
	var tests = []struct {
		input       string
		expectCreds Credentials
		expectErr   error
	}{
		{"username=user1\npassword=pass1", Credentials{"", "user1", "pass1"}, nil},
		{" \n username = user1 \n password = pass1 \n", Credentials{"", "user1", "pass1"}, nil},
		{"username = user1 \n junk \n password = pass1 \n", Credentials{"", "user1", "pass1"}, nil},
		{"USERNAME = user1 \n passed = pass1", Credentials{}, ErrMalformedSccCredFile},
		{"username= \n password = \n", Credentials{}, ErrMalformedSccCredFile},
	}

	for _, test := range tests {
		got, err := parseCredentials(strings.NewReader(test.input))
		if err != test.expectErr || got != test.expectCreds {
			t.Errorf("ParseCredentials() == %+v, %s, expected %+v, %s", got, err, test.expectCreds, test.expectErr)
		}
	}
}

func TestWriteReadDeleteSystem(t *testing.T) {
	CFG.FsRoot = t.TempDir()
	_, err := getCredentials()
	if err != ErrMissingCredentialsFile {
		t.Fatalf("Expected [%s], got [%s]", ErrMissingCredentialsFile, err)
	}
	if err := writeSystemCredentials("user1", "pass1"); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	c, err := getCredentials()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if c.Username != "user1" || c.Password != "pass1" {
		t.Errorf("Unexpected user1 and pass1. Got: %s and %s",
			c.Username, c.Password)
	}
	if err := removeSystemCredentials(); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	path := systemCredentialsFile()
	if fileExists(path) {
		t.Error("File was not deleted: ", path)
	}
}

func TestWriteReadDeleteService(t *testing.T) {
	CFG.FsRoot = t.TempDir()
	if err := writeSystemCredentials("user1", "pass1"); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if err := writeServiceCredentials("service1"); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	path := serviceCredentialsFile("service1")
	rc, err := readCredentials(path)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if rc.Username != "user1" || rc.Password != "pass1" {
		t.Errorf("Got: %s and %s, expected user1 and pass1", rc.Username, rc.Password)
	}
	if err := removeServiceCredentials("service1"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if fileExists(path) {
		t.Error("File was not deleted: ", path)
	}
}
