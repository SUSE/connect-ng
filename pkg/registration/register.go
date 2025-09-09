package registration

import (
	"encoding/json"

	"github.com/SUSE/connect-ng/pkg/connection"
)

type announceRequest struct {
	Hostname string `json:"hostname"`

	requestWithInformation
}

type announceResponse struct {
	Id       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register a system by using the given regcode. You also need to provide the
// hostname of the system plus extra system information that is to be bundled when registering
// a system.
// Additionally extraData can be supplied when extra information such as instance data or online at data is required
func Register(conn connection.Connection, regcode, hostname string, systemInformation SystemInformation, extraData ExtraData) (int, error) {
	reg := announceResponse{}
	payload := announceRequest{
		Hostname: hostname,
	}

	enrichWithSystemInformation(&payload.requestWithInformation, systemInformation)
	enrichErr := enrichWithExtraData(&payload.requestWithInformation, extraData)
	if enrichErr != nil {
		return 0, enrichErr
	}

	creds := conn.GetCredentials()

	request, buildErr := conn.BuildRequest("POST", "/connect/subscriptions/systems", payload)
	if buildErr != nil {
		return 0, buildErr
	}

	connection.AddRegcodeAuth(request, regcode)

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

	connection.AddSystemAuth(request, login, password)

	_, _, doErr := conn.Do(request)
	if doErr != nil {
		return doErr
	}

	return creds.SetLogin("", "")
}
