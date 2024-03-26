package credentials

import (
	"os"
	"strings"
	"testing"
)

func TestParseCredentials(t *testing.T) {
	var tests = []struct {
		input       string
		expectCreds Credentials
		expectErr   error
	}{
		{"username=user1\npassword=pass1", Credentials{"", "user1", "pass1", ""}, nil},
		{" \n username = user1 \n password = pass1 \nsystem_token=\n", Credentials{"", "user1", "pass1", ""}, nil},
		{"username = user1 \n junk \n password = pass1 \nsystem_token=1234", Credentials{"", "user1", "pass1", "1234"}, nil},
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
	_, err := GetCredentials()
	if err != ErrMissingCredentialsFile {
		t.Fatalf("Expected [%s], got [%s]", ErrMissingCredentialsFile, err)
	}
	if err := writeSystemCredentials("user1", "pass1", "1234"); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	c, err := GetCredentials()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if c.Username != "user1" || c.Password != "pass1" || c.SystemToken != "1234" {
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

func TestWriteCredentials(t *testing.T) {
	CFG.FsRoot = t.TempDir()
	if err := writeSystemCredentials("user1", "pass1", "1234"); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	expected := "username=user1\npassword=pass1\nsystem_token=1234\n"
	contents, _ := os.ReadFile(systemCredentialsFile())
	got := string(contents)
	if got != expected {
		t.Errorf("Expected %#v, got %#v", expected, got)
	}
}

func TestWriteCredentialsEmptyToken(t *testing.T) {
	CFG.FsRoot = t.TempDir()
	if err := writeSystemCredentials("user1", "pass1", ""); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	expected := "username=user1\npassword=pass1\n"
	contents, _ := os.ReadFile(systemCredentialsFile())
	got := string(contents)
	if got != expected {
		t.Errorf("Expected %#v, got %#v", expected, got)
	}
}

func TestWriteReadDeleteService(t *testing.T) {
	CFG.FsRoot = t.TempDir()
	if err := writeSystemCredentials("user1", "pass1", "1234"); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if err := writeServiceCredentials("service1"); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	path := serviceCredentialsFile("service1")
	rc, err := ReadCredentials(path)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if rc.Username != "user1" || rc.Password != "pass1" || rc.SystemToken != "" {
		t.Errorf("Got: %s, %s, %s. Expected user1, pass1, \"\"", rc.Username, rc.Password, rc.SystemToken)
	}
	if err := removeServiceCredentials("service1"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if fileExists(path) {
		t.Error("File was not deleted: ", path)
	}
}

func TestParseCurlrcCredentials(t *testing.T) {
	var tests = []struct {
		input       string
		expectCreds Credentials
		expectErr   error
	}{
		{"--proxy-user \"meuser1$:mepassord2%\"", Credentials{"", "meuser1$", "mepassord2%", ""}, nil},
		{"--proxy-user = \"meuser1$:mepassord2%\"", Credentials{"", "meuser1$", "mepassord2%", ""}, nil},
		{"proxy-user = \"meuser1$:mepassord2%\"", Credentials{"", "meuser1$", "mepassord2%", ""}, nil},
		{"proxy-user=\"meuser1$:mepassord2%\"", Credentials{"", "meuser1$", "mepassord2%", ""}, nil},
		{"", Credentials{}, ErrNoProxyCredentials},
	}

	for _, test := range tests {
		got, err := parseCurlrcCredentials(strings.NewReader(test.input))
		if err != test.expectErr || got != test.expectCreds {
			t.Errorf("parseCurlrcCredentials() == %+v, %s, expected %+v, %s, input: %s", got, err, test.expectCreds, test.expectErr, test.input)
		}
	}
}
