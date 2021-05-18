package connect

import (
	"encoding/json"
)

func GetActivations() ([]Activation, error) {
	urlSuffix := "connect/systems/activations"
	resp, err := DoGET(urlSuffix)
	if err != nil {
		return []Activation{}, err
	}
	var activations []Activation
	err = json.Unmarshal(resp, &activations)
	if err != nil {
		return []Activation{}, err
	}
	return activations, nil
}
