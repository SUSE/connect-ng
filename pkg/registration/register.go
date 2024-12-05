package registration

import "github.com/SUSE/connect-ng/pkg/connection"

// Register a system by using the given regcode. You also need to provide the
// hostname of the system plus extra info that is to be bundled when registering
// a system.
func Register(conn connection.Connection, regcode, hostname string, info any) error {
	return nil
}

// De-register the system pointed by the given authorized connection.
func Deregister(conn connection.Connection) error {
	return nil
}
