package connect

import (
	"fmt"
)

// Deregister deregisters the system
func Deregister() error {
	if !IsRegistered() {
		return ErrSystemNotRegistered
	}

	// TODO implement deregister
	return fmt.Errorf("Deregister not implemented yet")
}

// IsRegistered returns true if there is a valid credentials file
func IsRegistered() bool {
	_, err := getCredentials()
	return err == nil
}

// URLDefault returns true if using https://scc.suse.com
func URLDefault() bool {
	return CFG.BaseURL == defaultBaseURL
}

func printInformation(action string) {
	var server string
	if URLDefault() {
		server = "SUSE Customer Center"
	} else {
		server = "registration proxy " + CFG.BaseURL
	}
	if action == "register" {
		fmt.Printf(bold("Registering system to %s\n"), server)
	} else {
		fmt.Printf(bold("Deregistering system from %s\n"), server)
	}
	if CFG.FsRoot != "" {
		fmt.Println("Rooted at:", CFG.FsRoot)
	}
	if CFG.Email != "" {
		fmt.Println("Using E-Mail:", CFG.Email)
	}
}
