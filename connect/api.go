package connect

import (
	"encoding/json"
)

// GetActivations returns a map keyed by "Identifier/Version/Arch"
func GetActivations(creds Credentials) (map[string]Activation, error) {
	urlSuffix := "/connect/systems/activations"
	activeMap := make(map[string]Activation)
	resp, err := DoGET(creds, urlSuffix)
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
