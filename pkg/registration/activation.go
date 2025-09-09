package registration

import (
	"encoding/json"
	"time"

	"github.com/SUSE/connect-ng/pkg/connection"
)

// Activation holds information about a product activation.
// subscription information.
type Activation struct {
	Name             string    `json:"name"`
	Status           string    `json:"status"`
	RegistrationCode string    `json:"regcode"`
	Type             string    `json:"type"`
	StartsAt         time.Time `json:"starts_at"`
	ExpiresAt        time.Time `json:"expires_at"`

	Metadata *Metadata
	Product  *Product
}

// Returns the activation identified by the product's "triplet".
func (a *Activation) ToTriplet() string {
	p := a.Product
	return p.Identifier + "/" + p.Version + "/" + p.Arch
}

type activationResponse struct {
	Activation
	MetadataAndProduct activateResponse `json:"service"`
}

// Fetch all known product activations for this system. If there the system has not yet
// activated a product, it returns an empty array.
func FetchActivations(conn connection.Connection) ([]*Activation, error) {
	activations := []*activationResponse{}

	creds := conn.GetCredentials()
	login, password, credErr := creds.Login()

	if credErr != nil {
		return []*Activation{}, credErr
	}

	request, buildErr := conn.BuildRequest("GET", "/connect/systems/activations", nil)

	if buildErr != nil {
		return []*Activation{}, buildErr
	}

	connection.AddSystemAuth(request, login, password)

	_, response, doErr := conn.Do(request)
	if doErr != nil {
		return []*Activation{}, doErr
	}

	if err := json.Unmarshal(response, &activations); err != nil {
		return []*Activation{}, err
	}

	return unpackActivations(activations)
}

func unpackActivations(packed []*activationResponse) ([]*Activation, error) {
	activations := []*Activation{}

	for _, data := range packed {
		activation := &data.Activation
		activation.Product = &data.MetadataAndProduct.Product
		activation.Metadata = &Metadata{
			ID:            data.MetadataAndProduct.ID,
			URL:           data.MetadataAndProduct.URL,
			Name:          data.MetadataAndProduct.Name,
			ObsoletedName: data.MetadataAndProduct.ObsoletedName,
		}

		activations = append(activations, activation)
	}

	return activations, nil
}
