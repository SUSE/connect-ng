package connect

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	sampleLogin        = "SCC_a9b5e32370fb41e1baf99349f2780ae4"
	samplePassword     = "a3cd1331fb714e82"
	expectedDockerPath = "/home/test/.docker/config.json"
	expectedPodmanPath = "/var/run/1312/containers/auth.json"
	expectedRuntimeDir = "/var/run/1312/"
)

func testPathMatches(t *testing.T, path string) {
	if path != expectedDockerPath && path != expectedPodmanPath {
		t.Errorf("JSON path should be:\n `%s` or `%s` \n got: `%s`",
			expectedDockerPath,
			expectedPodmanPath,
			path)
	}
}

func mockCurrentUserHome(home string) {
	userHome = func() (string, error) {
		return home, nil
	}
}

func mockReadFile(t *testing.T, samplefile string) {
	readFile = func(path string) ([]byte, error) {
		testPathMatches(t, path)

		samplePath := filepath.Join("registry_auth", samplefile)
		return readTestFile(samplePath, t), nil
	}
}

func mockWriteFile(t *testing.T, matcherfile string) {
	writeFile = func(path string, content []byte, _ os.FileMode) error {
		testPathMatches(t, path)

		matcherPath := filepath.Join("registry_auth", matcherfile)
		expected := strings.Trim(string(readTestFile(matcherPath, t)), "\n")

		testContentMatches(t, expected, string(content))
		return nil
	}

}

func mockMkDirAll(t *testing.T) {
	mkDirAll = func(_ string, perm os.FileMode) error {
		if perm != 0777 {
			t.Log(fmt.Sprintf("mkdirall: %s is unlikely the right directory permission. Are you sure?", perm))
		}
		return nil
	}
}

func TestRegistryAuthSetupSuccessful(t *testing.T) {
	os.Setenv("XDG_RUNTIME_DIR", expectedRuntimeDir)
	mockMkDirAll(t)
	mockCurrentUserHome("/home/test")
	mockReadFile(t, "auth.json")
	mockWriteFile(t, "auth_updated.json")

	setupRegistryAuthentication(sampleLogin, samplePassword)
}

func TestRegistryAuthSetupReadFailed(t *testing.T) {
	os.Setenv("XDG_RUNTIME_DIR", expectedRuntimeDir)
	mockMkDirAll(t)
	mockCurrentUserHome("/home/test")
	mockWriteFile(t, "auth_write_single.json")

	readFile = func(path string) ([]byte, error) {
		return []byte{}, os.ErrNotExist
	}

	// Note: This will never fail, since it must not interrupt
	//       registration process
	setupRegistryAuthentication(sampleLogin, samplePassword)
}

func TestRegistryAuthSetupWriteDockerFailed(t *testing.T) {
	os.Setenv("XDG_RUNTIME_DIR", expectedRuntimeDir)
	mockCurrentUserHome("/home/test")
	mockReadFile(t, "empty_auth.json")

	writeFile = func(path string, content []byte, _ os.FileMode) error {
		// fail to docker config failed
		if path == expectedDockerPath {
			return fmt.Errorf("Permission denied")
		}

		expected := strings.Trim(string(readTestFile("registry_auth/auth_write_single.json", t)), "\n")
		testContentMatches(t, expected, string(content))
		return nil
	}

	setupRegistryAuthentication(sampleLogin, samplePassword)
}

func TestRegistryAuthRemoveSuccessful(t *testing.T) {
	os.Setenv("XDG_RUNTIME_DIR", expectedRuntimeDir)
	mockCurrentUserHome("/home/test")
	mockReadFile(t, "auth_updated.json")
	mockWriteFile(t, "auth.json")

	removeRegistryAuthentication(sampleLogin, samplePassword)
}

func TestRegistryAuthRemoveReadFailed(t *testing.T) {
	os.Setenv("XDG_RUNTIME_DIR", expectedRuntimeDir)
	mockCurrentUserHome("/home/test")

	readFile = func(_ string) ([]byte, error) {
		return []byte{}, os.ErrNotExist
	}

	writeFile = func(_ string, _ []byte, _ os.FileMode) error {
		fmt.Errorf("Expected writeFile to never be called")
		return nil
	}

	removeRegistryAuthentication(sampleLogin, samplePassword)
}
