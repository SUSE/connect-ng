package registration

import (
	"encoding/json"

	"github.com/SUSE/connect-ng/pkg/connection"
)

// Metadata holds all the data that is returned by activate/deactivate API calls
// which is not exactly tied to the Product struct. Note that by pairing a
// filled Metadata object and a Product object could give you, for example, a
// Zypper service.
type Metadata struct {
	// ID of the activation as given by SCC's API.
	ID int `json:"id"`

	// URL of the product activation so it can be used by other clients (e.g.
	// zypper).
	URL string `json:"url"`

	// Name of the product activation.
	Name string `json:"name"`

	// Extra name that is provided by SCC's APIs.
	ObsoletedName string `json:"obsoleted_service_name"`
}

type activateRequest struct {
	Identifier string `json:"identifier"`
	Version    string `json:"version"`
	Arch       string `json:"arch"`
	Regcode    string `json:"token,omitempty"`
}

type activateResponse struct {
	ID            int     `json:"id"`
	URL           string  `json:"url"`
	Name          string  `json:"name"`
	ObsoletedName string  `json:"obsoleted_service_name"`
	Product       Product `json:"product"`
}

// Activate a product by pairing an authorized connection (which contains the
// system at hand), plus the "triplet" being used to identify the desired
// product.
func Activate(conn connection.Connection, identifier, version, arch, regcode string) (*Metadata, *Product, error) {
	payload := activateRequest{
		Identifier: identifier,
		Version:    version,
		Arch:       arch,
		Regcode:    regcode,
	}
	creds := conn.GetCredentials()
	login, password, credErr := creds.Login()

	if credErr != nil {
		return nil, nil, credErr
	}

	request, buildErr := conn.BuildRequest("POST", "/connect/systems/products", payload)
	if buildErr != nil {
		return nil, nil, buildErr
	}

	connection.AddSystemAuth(request, login, password)

	response, doErr := conn.Do(request)
	if doErr != nil {
		return nil, nil, doErr
	}

	activation := activateResponse{}
	if err := json.Unmarshal(response, &activation); err != nil {
		return nil, nil, err
	}

	meta := Metadata{
		ID:            activation.ID,
		URL:           activation.URL,
		Name:          activation.Name,
		ObsoletedName: activation.ObsoletedName,
	}

	return &meta, &activation.Product, nil
}

// Deactivate a product by pairing an authorized connection (which contains the
// system at hand), plus the "triplet" being used to identify the product to be
// deactivated for the system.
func Deactivate(conn connection.Connection, identifier, version, arch string) (*Metadata, *Product, error) {
	payload := activateRequest{
		Identifier: identifier,
		Version:    version,
		Arch:       arch,
	}
	creds := conn.GetCredentials()
	login, password, credErr := creds.Login()

	if credErr != nil {
		return nil, nil, credErr
	}

	request, buildErr := conn.BuildRequest("DELETE", "/connect/systems/products", payload)
	if buildErr != nil {
		return nil, nil, buildErr
	}

	connection.AddSystemAuth(request, login, password)

	response, doErr := conn.Do(request)
	if doErr != nil {
		return nil, nil, doErr
	}

	deactivation := activateResponse{}
	if err := json.Unmarshal(response, &deactivation); err != nil {
		return nil, nil, err
	}

	meta := Metadata{
		ID:            deactivation.ID,
		URL:           deactivation.URL,
		Name:          deactivation.Name,
		ObsoletedName: deactivation.ObsoletedName,
	}

	return &meta, &deactivation.Product, nil
}
