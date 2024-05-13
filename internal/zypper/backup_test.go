package zypper

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// should align with the paths used in Backup()
var testPaths = []string{
	"etc/zypp/repos.d",
	"etc/zypp/credentials.d",
	"etc/zypp/services.d",
	"etc/products.d",
}

// should align with backup dir path in Backup()
var backupDir = "var/adm/backup/system-upgrade"

func sortedStringSlice(s []string) []string {
	sorted := make([]string, len(s))
	copy(sorted, s)
	slices.Sort(sorted)

	return sorted
}

func populateTestingRoot(t *testing.T, subPath string) {
	t.Helper()

	var filePerm os.FileMode = 0o644
	var dirPerm os.FileMode = 0o755
	data := []byte("content")

	// for each of the test paths, construct a path rooted under the
	// specified root directory, with the specified subPath appended,
	// and create the corresponding file, and any missing intermediary
	// directories
	for _, p := range testPaths {
		rootedFile := filepath.Join(zypperFilesystemRoot, p, subPath)
		rootedDir := filepath.Dir(rootedFile)

		err := os.MkdirAll(rootedDir, dirPerm)
		if err != nil {
			t.Fatalf("Failed to create testing dir %q: %s", rootedDir, err.Error())
		}

		err = os.WriteFile(rootedFile, data, filePerm)
		if err != nil {
			t.Fatalf("Failed to write test file %q: %s", rootedFile, err.Error())
		}
	}
}

func checkBackupCreated(t *testing.T) {

	assert := assert.New(t)

	tarballPath := filepath.Join(backupDir, "repos.tar.gz")
	scriptPath := filepath.Join(backupDir, "repos.sh")

	backupFiles := []string{
		tarballPath,
		scriptPath,
	}

	// verify that the backup files (tarball and restore script) were created
	for _, p := range backupFiles {
		rootedFile := filepath.Join(zypperFilesystemRoot, p)
		_, err := os.Stat(rootedFile)
		assert.NoError(err)
	}

	// verify that the restore script has expected entries
	rootedScript := filepath.Join(zypperFilesystemRoot, scriptPath)
	content, err := os.ReadFile(rootedScript)
	assert.NoError(err)
	rmPaths := []string{}
	var scriptTarball string
	for _, byteLine := range bytes.Split(content, []byte("\n")) {
		line := string(byteLine)

		// check for rm -rf lines and collect the associated paths
		prefix := "rm -rf " // should match zypper-restore.tmpl rm lines
		if strings.HasPrefix(line, prefix) {
			rmPath, _ := strings.CutPrefix(line, prefix)
			rmPaths = append(rmPaths, rmPath)
		}

		// check for a tar extract line, and remember the tarball path
		prefix = "tar xvf " // should match zypper-restore.tmp tar line
		if strings.HasPrefix(line, prefix) {
			scriptTarball = strings.Fields(line)[2]
		}
	}

	// sort the path slices to ensure valid comparison
	testPathsSorted := sortedStringSlice(testPaths)
	rmPathsSorted := sortedStringSlice(rmPaths)
	assert.Equal(testPathsSorted, rmPathsSorted)
	assert.Equal(tarballPath, scriptTarball)

	// verify that the tarball has expected entries
	rootedTarball := filepath.Join(zypperFilesystemRoot, tarballPath)
	cmd := exec.Command("tar", "tvaf", rootedTarball)
	tarList, err := cmd.Output()
	assert.NoError(err)

	// process tar listing output to extract list of top level directories
	// matching the test paths that should be included in the tarball.
	var tarDirs []string
	for _, tarLine := range bytes.Split(tarList, []byte("\n")) {
		line := string(tarLine)

		// skip blank lines
		if len(line) == 0 {
			continue
		}

		// skip non-directory entries
		if !strings.HasPrefix(line, "d") {
			continue
		}

		// extract the last field of the line and strip off trailing "/"
		lineFields := strings.Fields(line)
		dirPath := strings.TrimRight(lineFields[len(lineFields)-1], "/")

		// check if directory entry is a test path subdirectory
		var found bool
		for _, tp := range testPaths {
			if strings.Contains(dirPath, tp) && dirPath != tp {
				found = true
				break
			}
		}

		// ignore test path subdirectories
		if !found {
			tarDirs = append(tarDirs, dirPath)
		}
	}

	// sort the tarDirs list to ensure valid comparison
	tarDirsSorted := sortedStringSlice(tarDirs)
	assert.Equal(testPathsSorted, tarDirsSorted)
}

func checkRestoreState(t *testing.T, expected, notExpected string) {
	assert := assert.New(t)

	// ensure that the expected file exists in each test dir, and that the
	// notExpected file has been removed.
	for _, p := range testPaths {
		expectedPath := filepath.Join(zypperFilesystemRoot, p, expected)
		notExpectedPath := filepath.Join(zypperFilesystemRoot, p, notExpected)

		// expected files were created before backup was made
		_, err := os.Stat(expectedPath)
		assert.NoError(err)

		// notExpected files were created after backup was made and
		// should have been removed by the restore
		_, err = os.Stat(notExpectedPath)
		if assert.Error(err) {
			assert.ErrorContains(err, "no such file or directory")
		}
	}
}

func TestBackupAndRestore(t *testing.T) {
	assert := assert.New(t)
	expected := filepath.Join("back", "this", "up")
	notExpected := filepath.Join("not", "backed", "up")
	zypperFilesystemRoot = t.TempDir()

	// populate testing tree with required directories, each containing
	// the expected file
	populateTestingRoot(t, expected)

	// trigger a backup and verify that the backup was created as expected
	err := Backup()
	assert.NoError(err)
	checkBackupCreated(t)

	// now add the notExpected file to each of the required directories
	populateTestingRoot(t, notExpected)

	// trigger a restore and verify that notExpected files are not present
	err = Restore()
	assert.NoError(err)
	checkRestoreState(t, expected, notExpected)
}
