package credentials

import "testing"

func CreateTestCredentials(username, password string, fsRoot string, t *testing.T) {
	t.Helper()
	if username == "" {
		username = "test"
	}
	if password == "" {
		password = "test"
	}
	err := CreateCredentials(username, password, "", SystemCredentialsPath(fsRoot))
	if err != nil {
		t.Fatal(err)
	}
}
