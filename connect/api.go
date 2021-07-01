package connect

import (
	"encoding/json"
	"net/http"
)

const (
	// APIVersion is the SCC API version
	APIVersion = "v4"
)

// UpToDate Checks if API endpoint is up-to-date,
// useful when dealing with RegistrationProxy errors
func UpToDate() bool {
	// REVIST 404 case - see original
	// Should fail in any case. 422 error means that the endpoint is there and working right
	_, err := callHTTP("GET", "/connect/repositories/installer", nil, nil, authNone)
	if err == nil {
		return false
	}
	if ae, ok := err.(APIError); ok {
		if ae.Code == http.StatusUnprocessableEntity {
			return true
		}
	}
	return false
}

// GetActivations returns a map keyed by "Identifier/Version/Arch"
func GetActivations() (map[string]Activation, error) {
	activeMap := make(map[string]Activation)
	resp, err := callHTTP("GET", "/connect/systems/activations", nil, nil, authSystem)
	if err != nil {
		return activeMap, err
	}
	var activations []Activation
	err = json.Unmarshal(resp, &activations)
	if err != nil {
		return activeMap, err
	}
	for _, activation := range activations {
		activeMap[activation.toTriplet()] = activation
	}
	return activeMap, nil
}

func GetProduct(productQuery Product) (Product, error) {
	resp, err := callHTTP("GET", "/connect/systems/products", nil, productQuery.toQuery(), authSystem)
	remoteProduct := Product{}
	if err != nil {
		return remoteProduct, err
	}
	err = json.Unmarshal(resp, &remoteProduct)
	return remoteProduct, err
}

func upgradeProduct(product Product) (Service, error) {
	// NOTE: this can add some extra attributes to json payload which
	//       seem to be safely ignored by the API.
	payload, err := json.Marshal(product)
	remoteService := Service{}
	if err != nil {
		return remoteService, err
	}
	resp, err := callHTTP("PUT", "/connect/systems/products", payload, nil, authSystem)
	if err != nil {
		return remoteService, err
	}
	err = json.Unmarshal(resp, &remoteService)
	return remoteService, err
}
