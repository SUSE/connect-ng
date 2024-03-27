package connect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SUSE/connect-ng/internal/credentials"
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
	err := credentials.CreateCredentials(username, password, "", credentials.SystemCredentialsPath(CFG.FsRoot))
	if err != nil {
		t.Fatal(err)
	}
}

func testContentMatches(t *testing.T, expected string, got string) {
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
