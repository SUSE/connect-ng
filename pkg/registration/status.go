package registration

import (
	"net/http"

	"github.com/SUSE/connect-ng/internal/util"
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

	// The profile cache should be cleared.
	ClearCache

	// Set when an error occurred
	Unknown
)

type statusRequest struct {
	Hostname string `json:"hostname"`

	requestWithInformation
}

// Returns the registration status for the system pointed by the authorized
// connection.
func Status(conn connection.Connection, hostname string, systemInformation SystemInformation, profiles DataProfiles, extraData ExtraData) (StatusCode, error) {
	payload := statusRequest{
		Hostname: hostname,
	}

	enrichWithSystemInformation(&payload.requestWithInformation, systemInformation)
	payload.DataProfiles = profiles
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

	util.Debug.Println("registration payload : ", payload)
	util.Debug.Println("registration request : ", request)
	code, response, doErr := conn.Do(request)
	if doErr != nil {
		util.Debug.Println("registration.Status  code. doErr: ", code, doErr)
		if code == http.StatusUnauthorized {
			return Unregistered, nil
		}
		return Unknown, doErr
	}

	util.Debug.Println("registration.Status response: ", response)
	if code == http.StatusResetContent {
		return ClearCache, nil
	}
	return Registered, nil
}
