package registration

import (
	"github.com/SUSE/connect-ng/pkg/connection"
)

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
	Hostname string `json:"hostname"`

	requestWithInformation
}

// Returns the registration status for the system pointed by the authorized
// connection.
func Status(conn connection.Connection, hostname string, systemInformation SystemInformation, extraData ExtraData) (StatusCode, error) {
	payload := statusRequest{
		Hostname: hostname,
	}

	enrichWithSystemInformation(&payload.requestWithInformation, systemInformation)
	enrichErr := enrichWithExtraData(&payload.requestWithInformation, extraData)
	if enrichErr != nil {
		return 0, enrichErr
	}

	request, buildErr := conn.BuildRequest("PUT", "/connect/systems", payload)
	if buildErr != nil {
		return Unknown, buildErr
	}

	login, password, credErr := conn.GetCredentials().Login()
	if credErr != nil {
		return Unknown, credErr
	}

	connection.AddSystemAuth(request, login, password)

	_, doErr := conn.Do(request)
	if doErr != nil {
		return Unregistered, nil
	}

	return Registered, nil
}
