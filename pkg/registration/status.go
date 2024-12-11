package registration

import "github.com/SUSE/connect-ng/pkg/connection"

// Enum being used to report on the different status scenarios for a given
// connection.
type StatusCode int

const (
	// System has been registered.
	Registered StatusCode = iota

	// System is not registered yet.
	Unregistered

	// System has an expired subscription.
	Expired
)

// Returns the registration status for the system pointed by the authorized
// connection.
func Status(conn connection.Connection) (StatusCode, error) {
	return Registered, nil
}
