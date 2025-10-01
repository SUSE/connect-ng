package registration

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"

	"github.com/SUSE/connect-ng/pkg/connection"
)

// Encodes the offline requests product which can be activated
type OfflineRequestProductTriplet struct {
	Identifier string `json:"identifier"`
	Version    string `json:"version"`
	Arch       string `json:"arch"`
}

// This structure is the opaque version of the base64 encoded offline registration request.
// It includes all necessary and optional information required to created and activate a system
// and retrieve an offline registration certificate either via the API [registration.RegisterWithOfflineRequest] or
// via the [SCC UI]: https://scc.suse.com/register-offline/rancher
type OfflineRequest struct {
	Product OfflineRequestProductTriplet `json:"product"`

	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`

	SystemInformation SystemInformation `json:"system_information,omitempty"`
}

// Sets credentials to be included in the offline request. This method fails if credentials are not available.
func (req *OfflineRequest) SetCredentials(creds connection.Credentials) error {
	login, password, readErr := creds.Login()
	if readErr != nil {
		return readErr
	}

	req.Login = login
	req.Password = password

	return nil
}

// Marshal and encode the offline request into its base64 representation.
func (req *OfflineRequest) Base64Encoded() (io.Reader, error) {
	data, jsonErr := json.Marshal(req)

	if jsonErr != nil {
		return nil, jsonErr
	}
	blob := base64.StdEncoding.EncodeToString(data)
	reader := strings.NewReader(strings.TrimSpace(blob))

	return reader, nil
}

// Builds an offline registration request with it respective required attributes.
// See [OfflineRequest.SetCredentials] for optional attributes.
func BuildOfflineRequest(identifier, version, arch string, systemInformation SystemInformation) *OfflineRequest {
	return &OfflineRequest{
		Product: OfflineRequestProductTriplet{
			Identifier: identifier,
			Version:    version,
			Arch:       arch,
		},
		SystemInformation: systemInformation,
	}
}

// Create an offline certificate using the API. This needs a connection to the registration API. See [registration.OfflineCertificateFrom] how to create an offline registration from a reader instead the API.
func RegisterWithOfflineRequest(conn connection.Connection, regcode string, offlineRequest *OfflineRequest) (*OfflineCertificate, error) {
	body, requestErr := offlineRequest.Base64Encoded()
	if requestErr != nil {
		return nil, requestErr
	}

	request, buildErr := conn.BuildRequestRaw("POST", "/connect/subscriptions/offline-register", body)
	if buildErr != nil {
		return nil, buildErr
	}

	connection.AddRegcodeAuth(request, regcode)
	// Replace `application/json` with the appropriate header to make sure
	// the API is not parsing the request blob as compressed JSON
	request.Header.Set("Content-Type", "text/plain")

	_, response, doErr := conn.Do(request)
	if doErr != nil {
		return nil, doErr
	}
	reader := bytes.NewReader(response)

	certificate, parseErr := OfflineCertificateFrom(reader, true)
	if parseErr != nil {
		return nil, parseErr
	}

	return certificate, nil
}
