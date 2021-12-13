package connect

import (
	"os"
	"path/filepath"
	"testing"
)

func readTestFile(name string, t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("../../testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func createTestCredentials(username, password string, t *testing.T) {
	t.Helper()
	if username == "" {
		username = "test"
	}
	if password == "" {
		password = "test"
	}
	CFG.FsRoot = t.TempDir()
	err := writeSystemCredentials(username, password)
	if err != nil {
		t.Fatal(err)
	}
}
