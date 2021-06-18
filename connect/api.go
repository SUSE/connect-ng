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
	_, err := callHTTP("GET", "/connect/repositories/installer", nil, nil, Credentials{})
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
func GetActivations(creds Credentials) (map[string]Activation, error) {
	activeMap := make(map[string]Activation)
	resp, err := callHTTP("GET", "/connect/systems/activations", nil, nil, creds)
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
