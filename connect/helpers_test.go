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
