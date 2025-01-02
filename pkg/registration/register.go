package registration

import (
	"encoding/json"

	"github.com/SUSE/connect-ng/pkg/connection"
)

type announceRequest struct {
	Hostname          string `json:"hostname"`
	SystemInformation any    `json:"hwinfo"`
}

type announceResponse struct {
	Id       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register a system by using the given regcode. You also need to provide the
// hostname of the system plus extra info that is to be bundled when registering
// a system.
func Register(conn connection.Connection, regcode, hostname string, systemInformation any) (int, error) {
	reg := announceResponse{}
	payload := announceRequest{
		Hostname:          hostname,
		SystemInformation: systemInformation,
	}
	creds := conn.GetCredentials()

	request, buildErr := conn.BuildRequest("POST", "/connect/subscriptions/systems", payload)
	if buildErr != nil {
		return 0, buildErr
	}

	connection.AuthByRegcode(request, regcode)

	_, response, doErr := conn.Do(request)
	if doErr != nil {
		return 0, doErr
	}

	if err := json.Unmarshal(response, &reg); err != nil {
		return 0, err
	}

	credErr := creds.SetLogin(reg.Login, reg.Password)

	return reg.Id, credErr
}

// De-register the system pointed by the given authorized connection.
func Deregister(conn connection.Connection) error {
	creds := conn.GetCredentials()
	request, buildErr := conn.BuildRequest("DELETE", "/connect/systems", nil)
	if buildErr != nil {
		return buildErr
	}

	login, password, credErr := creds.Login()
	if credErr != nil {
		return credErr
	}

	connection.AuthBySystemCredentials(request, login, password)

	_, _, doErr := conn.Do(request)
	if doErr != nil {
		return doErr
	}

	return creds.SetLogin("", "")
}
