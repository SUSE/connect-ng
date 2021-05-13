package connect

import (
	"errors"
	"log"
)

func GetActivations() []Activation {
	urlSuffix := "connect/systems/activations"
	resp, err := DoGET(urlSuffix)
	if err != nil {
		// A missing credentials file just means the system is
		// not registered. Don't print an error in this case.
		if !errors.Is(err, ErrNoCredentialsFile) {
			log.Println(err)
		}
		return []Activation{}
	}
	return ParseJSON(resp)
}
