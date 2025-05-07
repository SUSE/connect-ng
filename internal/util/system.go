package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"syscall"
)

const (
	// From <linux/mount.h>, <bits/statvfs.h>, etc.
	ST_RDONLY = 0x1

	// From <linux/magic.h>
	//
	// NOTE: this will not work on 32-bit systems (i.e. armv7 in 32-bit mode).
	// If we ever consider supporting these systems, then we would need to
	// change this.
	BTRFS_SUPER_MAGIC = 0x9123683E

	// See: https://cs.opensource.google/go/go/+/master:src/internal/goos/zgoos_linux.go
	LINUX_IDENTIFIER = "linux"
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
var Execute = func(cmd []string, validExitCodes []int) ([]byte, error) {
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
	if err != nil && !slices.Contains(validExitCodes, exitCode) {
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

var FileExists = func(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

var ReadFile = func(file string) []byte {
	if FileExists(file) {
		b, _ := os.ReadFile(file)
		return b
	}
	return []byte{}
}

var ReadFileString = func(file string) string {
	if FileExists(file) {
		b, _ := os.ReadFile(file)
		return string(b)
	}
	return ""
}

var ExecutableExists = func(path string) bool {
	_, err := exec.LookPath(path)
	return err == nil
}

func RemoveFile(path string) error {
	Debug.Print("Removing file: ", path)
	if !FileExists(path) {
		return nil
	}
	return os.Remove(path)
}

func IsRootFSWritable() bool {
	_, err := Execute([]string{"test", "-w", "/"}, []int{0})
	return err == nil
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

	if (statfs.Flags & ST_RDONLY) == ST_RDONLY {
		// Just like zypper, we will assume that a BTRFS file system with the
		// `transactional-update` binary installed is a transactional server.
		_, err := os.Stat("/usr/sbin/transactional-update")
		if statfs.Type == BTRFS_SUPER_MAGIC && err == nil {
			// The user did not use the 'root' flag from the CLI: this is not
			// `transactional-update` calling SUSEConnect but rather a user
			// directly which is dicouraged.
			if root == "" {
				return errors.New("This is a transactional system, please use `transactional-update register` to manage your product activations")
			}

			return nil
		}
		// The root is read only, we cannot write in there.
		return fmt.Errorf("`%v` is mounted as `read-only`. Aborting", path)
	}
	return nil
}

var CurrentUser = func() string {
	u, err := user.Current()
	if err != nil {
		Debug.Fatal("cannot determine the current user")
		return ""
	}
	return u.Username
}

// Returns the home directory from the given login name by reading
// `/etc/passwd`. If an error occurs (e.g. `/etc/passwd` could not be read),
// then an empty string is simply returned.
func homeFromUser(name string) string {
	b := ReadFile("/etc/passwd")
	if len(b) == 0 {
		return ""
	}

	re, err := regexp.Compile("(?m)^" + name + ":(.*)")
	if err != nil {
		return ""
	}

	results := re.FindSubmatch(b)
	if len(results) == 2 {
		entries := strings.Split(string(results[0]), ":")
		if len(entries) >= 6 {
			return entries[5]
		}
	}
	return ""
}

func CurrentHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Sadly Go only reads from "$HOME" in order to find the user's home
		// directory and does not want to go any further if that environment
		// variable is not defined. Let's try with `homeFromUser` which tries to
		// pick up this value by reading the current user's /etc/passwd entry.
		if u := CurrentUser(); u == "" {
			// If we are still not able to detect it, let's assume "/" and cross
			// fingers.
			home = "/"
		} else {
			home = homeFromUser(u)
		}
	}
	return home
}

// Returns an error if the current platform is not Linux.
//
// NOTE: Android is considered a different platform than "Linux".
func EnsureLinux() error {
	if platform := runtime.GOOS; platform != LINUX_IDENTIFIER {
		return fmt.Errorf("platform '%v' is not supported. Use Linux instead", platform)
	}
	return nil
}

// Ensure that this is a Linux box with Zypper installed.
func EnsureZypper() error {
	if err := EnsureLinux(); err != nil {
		return err
	}

	if !ExecutableExists("zypper") {
		return fmt.Errorf("'zypper' is not installed. Make sure it's on your PATH")
	}
	return nil
}
