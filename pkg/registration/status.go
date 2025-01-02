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

	// Set when an error occurred
	Unknown
)

type statusRequest struct {
	Hostname          string `json:"hostname"`
	SystemInformation any    `json:"hwinfo"`
}

// Returns the registration status for the system pointed by the authorized
// connection.
func Status(conn connection.Connection, hostname string, systemInformation any) (StatusCode, error) {
	payload := statusRequest{
		Hostname:          hostname,
		SystemInformation: systemInformation,
	}

	request, buildErr := conn.BuildRequest("PUT", "/connect/systems", payload)
	if buildErr != nil {
		return Unknown, buildErr
	}

	login, password, credErr := conn.GetCredentials().Login()
	if credErr != nil {
		return Unknown, credErr
	}

	connection.AuthBySystemCredentials(request, login, password)

	_, _, doErr := conn.Do(request)
	if doErr != nil {
		return Unregistered, nil
	}

	return Registered, nil
}
