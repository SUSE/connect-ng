package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func CleanupPolutedFilesystem() {
	fmt.Printf("[cleanup] Cleanup /etc/zypp/{credentials.d, services.d, repos.d}/*...\n")
	removeInnerRecursive("/etc/zypp/credentials.d")
	removeInnerRecursive("/etc/zypp/services.d")
	removeInnerRecursive("/etc/zypp/repos.d")
}

func removeInnerRecursive(dir string) {
	matches, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return
	}
	for _, match := range matches {
		err = os.RemoveAll(match)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error removing contents of %s: %v\n", dir, err)
			return
		}
	}
}

func RemoveFile(t *testing.T, path string) {
	err := os.Remove(path)
	if err != nil {
		assert.FailNow(t, "Failed to remove file", "Error removing file %s: %v", path, err)
	}
}

func FriendlyNameToServiceName(friendlyName string) string {
	// Duplicates the naming convention created by SCC.
	// Check the service model in SCC for more information
	return strings.ReplaceAll(friendlyName, " ", "_")
}

func FriendlyNameToCredentialsName(friendlyName string) string {
	return FriendlyNameToServiceName(friendlyName)
}

func TryCurlrcCleanup() {
	home := os.Getenv("HOME")
	if home == "" {
		panic("empty HOME")
	}

	_ = os.Remove(filepath.Join(home, ".curlrc"))
}
