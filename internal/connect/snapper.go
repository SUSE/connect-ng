package connect

import (
	"bufio"
	"strconv"
	"strings"

	"github.com/SUSE/connect-ng/internal/util"
)

const (
	snapperPath = "/usr/bin/snapper"
)

func createSnapshot(snapshotType, desc string, args []string) (int, error) {
	cmd := []string{snapperPath, "create", "--type", snapshotType,
		"--cleanup-algorithm", "number", "--print-number",
		"--userdata", "important=yes", "--description", desc}
	cmd = append(cmd, args...)
	util.QuietOut.Printf("\nExecuting '%s'\n\n", strings.Join(cmd, " "))
	output, err := util.Execute(cmd, []int{0})
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(output))
}

// CreatePreSnapshot creates "pre" type snapshot
func CreatePreSnapshot() (int, error) {
	return createSnapshot("pre", "before online migration", []string{})
}

// CreatePostSnapshot creates "post" type snapshot for given preSnapshot
func CreatePostSnapshot(preSnapshot int) (int, error) {
	return createSnapshot("post", "after online migration",
		[]string{"--pre-number", strconv.Itoa(preSnapshot)})
}

// IsSnapperConfigured checks if snapper is properly configured
func IsSnapperConfigured() bool {
	output, err := util.Execute([]string{snapperPath, "--no-dbus", "list-configs"}, []int{0})
	if err != nil {
		return false
	}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "root ") {
			return true
		}
	}
	return false
}
