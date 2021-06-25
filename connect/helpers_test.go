package connect

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func readTestFile(name string, t *testing.T) []byte {
	data, err := os.ReadFile(filepath.Join("../testdata", name))
	if err != nil {
		_, filename, num, _ := runtime.Caller(1)
		t.Fatalf("\n%s:%d: %s", filepath.Base(filename), num, err)
	}
	return data
}

func createTestCredentials(username, password string, t *testing.T) {
	if username == "" {
		username = "test"
	}
	if password == "" {
		password = "test"
	}
	CFG.FsRoot = t.TempDir()
	err := writeSystemCredentials(username, password)
	if err != nil {
		_, filename, num, _ := runtime.Caller(1)
		t.Fatalf("\n%s:%d: %s", filepath.Base(filename), num, err)
	}
}
