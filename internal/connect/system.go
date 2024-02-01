package connect

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const (
	// From <linux/mount.h>, <bits/statvfs.h>, etc.
	ST_RDONLY = 0x1

	// From <linux/magic.h>
	BTRFS_SUPER_MAGIC = 0x9123683E
)

var systemEcho bool

// SetSystemEcho toggles piping of executed command's outputs to stdout/stderr
// returns true if it was enabled before, false otherwise
func SetSystemEcho(v bool) bool {
	prev := systemEcho
	systemEcho = v
	return prev
}

// Assign function for running external commands to a variable so it can be mocked by tests.
var execute = func(cmd []string, validExitCodes []int) ([]byte, error) {
	Debug.Print("Executing: ", cmd)
	var stderr, stdout bytes.Buffer
	comm := exec.Command(cmd[0], cmd[1:]...)
	if systemEcho {
		comm.Stdout = io.MultiWriter(os.Stdout, &stdout)
		comm.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		comm.Stdout = &stdout
		comm.Stderr = &stderr
	}
	comm.Env = append(os.Environ(), "LC_ALL=C")
	err := comm.Run()
	exitCode := comm.ProcessState.ExitCode()
	Debug.Printf("Return code: %d\n", exitCode)
	if stdout.Len() > 0 {
		Debug.Print("Output: ", stdout.String())
	}
	if stderr.Len() > 0 {
		Debug.Print("Error: ", stderr.String())
	}
	// TODO Ruby version also checks stderr for "ABORT request"
	if err != nil && !containsInt(validExitCodes, exitCode) {
		output := stderr.Bytes()
		// zypper with formatter option writes to stdout instead of stderr
		if len(output) == 0 {
			output = stdout.Bytes()
		}
		output = bytes.TrimSuffix(output, []byte("\n"))
		ee := ExecuteError{Commmand: cmd, ExitCode: exitCode, Output: output, Err: err}
		return nil, ee
	}
	out := stdout.Bytes()
	out = bytes.TrimSuffix(out, []byte("\n"))
	return out, nil
}

func containsInt(s []int, i int) bool {
	for _, e := range s {
		if e == i {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func removeFile(path string) error {
	Debug.Print("Removing file: ", path)
	if !fileExists(path) {
		return nil
	}
	return os.Remove(path)
}

func isRootFSWritable() bool {
	_, err := execute([]string{"test", "-w", "/"}, []int{zypperOK})
	return err == nil
}

// Cleanup removes system credentials and installed services
func Cleanup() error {
	err := removeSystemCredentials()
	if err != nil {
		return err
	}

	// remove all suse services from zypper
	services, err := InstalledServices()
	if err != nil {
		return err
	}

	for _, service := range services {
		// NOTE: this check might not work correctly with SMT depending
		//       on the configuration (e.g. listen on https but API
		//       returns URL with http).
		if !strings.Contains(service.URL, CFG.BaseURL) {
			Debug.Printf("%s not in %s\n", CFG.BaseURL, service.URL)
			continue
		}
		if err := removeService(service.Name); err != nil {
			return err
		}

	}
	return nil
}

// UpdateCertificates runs system certificate update command
func UpdateCertificates() error {
	cmd := []string{"/usr/sbin/update-ca-certificates"}
	_, err := execute(cmd, []int{0})
	if err != nil {
		return err
	}
	// reload CA certs in Go
	return ReloadCertPool()
}

// ReadOnlyFilesystem returns an error if the given root path contains a
// read-only mount point or if the system should actually be managed through
// `transactional-update`. Otherwise it just returns nil. Note that if the given
// root path is empty, then "/" is assumed.
func ReadOnlyFilesystem(root string) error {
	path := root
	if path == "" {
		path = "/"
	}
	statfs := &syscall.Statfs_t{}
	if err := syscall.Statfs(path, statfs); err != nil {
		return fmt.Errorf("Checking whether %v is mounted read-only failed: %v", path, err)
	}

	// Just like zypper, we will assume that a BTRFS file system with the
	// `transactional-update` binary installed is a transactional server.
	_, err := os.Stat("/usr/sbin/transactional-update")
	if statfs.Type == BTRFS_SUPER_MAGIC && err == nil {
		// The user did not use the 'root' flag from the CLI: this is not
		// `transactional-update` calling SUSEConnect but rather a user
		// directly. This is not allowed.
		if root == "" {
			return errors.New("This is a transactional-server, please use `transactional-update register` to manage your product activations")
		}
	}

	// The root is read only, we cannot write in there.
	if (statfs.Flags & ST_RDONLY) == ST_RDONLY {
		return fmt.Errorf("`%v` is mounted as `read-only`. Aborting", path)
	}

	return nil
}
