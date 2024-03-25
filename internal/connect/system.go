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
