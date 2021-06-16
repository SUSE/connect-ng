package connect

import (
	"fmt"
)

// Deregister deregisters the system
func Deregister() error {
	if !isRegistered() {
		return ErrSystemNotRegistered
	}

	// TODO implement deregister
	return fmt.Errorf("Deregister not implemented yet")
}

func isRegistered() bool {
	return fileExists(defaulCredPath)
}
