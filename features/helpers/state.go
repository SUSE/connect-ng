package helpers

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/SUSE/connect-ng/internal/zypper"
)

func TrySUSEConnectCleanup() {
	fmt.Printf("[cleanup] Try running suseconnect --cleanup...\n")
	args := []string{}

	// Add --root flag if custom root is set
	if zypper.GetFilesystemRoot() != "/" {
		args = append(args, "--root", zypper.GetFilesystemRoot())
	}

	args = append(args, "--cleanup")
	cmd := exec.Command("suseconnect", args...)
	err := cmd.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, " => Failed to run suseconnect --cleanup. Skipping..\n")
	}
}

func TrySUSEConnectDeregister() {
	fmt.Printf("[cleanup] Try running suseconnect --de-register...\n")
	args := []string{}

	// Add --root flag if custom root is set
	if zypper.GetFilesystemRoot() != "/" {
		args = append(args, "--root", zypper.GetFilesystemRoot())
	}

	args = append(args, "--de-register")
	cmd := exec.Command("suseconnect", args...)
	err := cmd.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, " => Failed to run suseconnect --de-register. Skipping..\n")
	}
}
