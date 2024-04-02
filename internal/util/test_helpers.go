package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func ReadTestFile(name string, t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("../../testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestContentMatches(t *testing.T, expected string, got string) {
	if expected != got {
		message := []string{"write: Expected content to match:",
			"---",
			"%s",
			"---",
			"but got:",
			"---",
			"%s",
			"---"}
		t.Errorf(strings.Join(message, "\n"), expected, got)
	}
}
