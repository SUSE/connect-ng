package credentials

import (
	"os"
	"strings"
	"testing"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/assert"
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
	fsRoot := t.TempDir()
	sysCredsPath := SystemCredentialsPath(fsRoot)
	_, err := ReadCredentials(sysCredsPath)
	if err != ErrMissingCredentialsFile {
		t.Fatalf("Expected [%s], got [%s]", ErrMissingCredentialsFile, err)
	}
	if err := CreateCredentials("user1", "pass1", "1234", sysCredsPath); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	c, err := ReadCredentials(sysCredsPath)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if c.Username != "user1" || c.Password != "pass1" || c.SystemToken != "1234" {
		t.Errorf("Expected user1 and pass1. Got: %s and %s",
			c.Username, c.Password)
	}
	if err := util.RemoveFile(sysCredsPath); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	// TODO : Move to util package
	// Looks like we're verifying if util.RemoveFile works correctly here.
	// Instead, we should directly write a test for this method within the util package
	if util.FileExists(sysCredsPath) {
		t.Error("File was not deleted: ", sysCredsPath)
	}
}

func TestWriteCredentials(t *testing.T) {
	fsRoot := t.TempDir()
	sysCredsPath := SystemCredentialsPath(fsRoot)
	if err := CreateCredentials("user1", "pass1", "1234", sysCredsPath); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	expected := "username=user1\npassword=pass1\nsystem_token=1234\n"
	contents, _ := os.ReadFile(sysCredsPath)
	got := string(contents)
	if got != expected {
		t.Errorf("Expected %#v, got %#v", expected, got)
	}
}

func TestWriteCredentialsEmptyToken(t *testing.T) {
	fsRoot := t.TempDir()
	sysCredsPath := SystemCredentialsPath(fsRoot)
	if err := CreateCredentials("user1", "pass1", "", sysCredsPath); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	expected := "username=user1\npassword=pass1\n"
	contents, _ := os.ReadFile(sysCredsPath)
	got := string(contents)
	if got != expected {
		t.Errorf("Expected %#v, got %#v", expected, got)
	}
}

func TestWriteReadDeleteService(t *testing.T) {
	fsRoot := t.TempDir()
	sysCredsPath := SystemCredentialsPath(fsRoot)
	if err := CreateCredentials("user1", "pass1", "1234", sysCredsPath); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	systemCreds, err := ReadCredentials(sysCredsPath)
	if err != nil {
		t.Fatalf("Unable to read system credentials: %s", err)
	}
	serviceCredPath := ServiceCredentialsPath("service1", fsRoot)
	if err := CreateCredentials(systemCreds.Username, systemCreds.Password, "", serviceCredPath); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	rc, err := ReadCredentials(serviceCredPath)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if rc.Username != "user1" || rc.Password != "pass1" || rc.SystemToken != "" {
		t.Errorf("Got: %s, %s, %s. Expected user1, pass1, \"\"", rc.Username, rc.Password, rc.SystemToken)
	}
	if err := util.RemoveFile(serviceCredPath); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	// TODO : Move to util package
	if util.FileExists(serviceCredPath) {
		t.Error("File was not deleted: ", serviceCredPath)
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

func TestCurlrcCredentialsPath(t *testing.T) {
	assert := assert.New(t)

	// We have to manipulate the path for this test to work.
	home := os.Getenv("HOME")
	err := os.Unsetenv("HOME")
	if err != nil {
		t.Fatal("could not setup test")
	}
	defer os.Setenv("HOME", home)

	// Mocking up important functions

	util.CurrentUser = func() string { return "user" }
	util.ReadFile = func(path string) []byte {
		assert.Equal("/etc/passwd", path)
		return util.ReadTestFile("credentials/passwd.txt", t)
	}

	// The actual test :)

	assert.Equal("/home/user/.curlrc", CurlrcCredentialsPath(), "bad curlrc credentials path")
}
