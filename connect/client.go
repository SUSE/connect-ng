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

// IsRegistered returns true if the system credentials file exists
func IsRegistered() bool {
	return fileExists(defaulCredPath)
}

// URLDefault returns true if using https://scc.suse.com
func URLDefault() bool {
	return CFG.BaseURL == defaultBaseURL
}
