package helpers

import (
	"fmt"
	"os"
	"os/exec"
)

func TrySUSEConnectCleanup() {
	fmt.Printf("[cleanup] Try running suseconnect --cleanup...\n")
	cmd := exec.Command("suseconnect", "--cleanup")
	err := cmd.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, " => Failed to run suseconnect --cleanup. Skipping..\n")
	}
}

func TrySUSEConnectDeregister() {
	fmt.Printf("[cleanup] Try running suseconnect --de-register...\n")
	cmd := exec.Command("suseconnect", "--de-register")
	err := cmd.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, " => Failed to run suseconnect --de-register. Skipping..\n")
	}
}
