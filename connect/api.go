package connect

import (
	"encoding/json"
	"errors"
)

func GetActivations() []Activation {
	urlSuffix := "connect/systems/activations"
	resp, err := DoGET(urlSuffix)
	if err != nil {
		// A missing credentials file just means the system is
		// not registered. Don't print an error in this case.
		if !errors.Is(err, ErrNoCredentialsFile) {
			Error.Println(err)
		}
		return []Activation{}
	}
	var activations []Activation
	err = json.Unmarshal(resp, &activations)
	if err != nil {
		Error.Fatal(err)
	}
	return activations
}
