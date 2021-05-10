package connect

import (
	"log"
)

func GetActivations() []Activation {
	urlSuffix := "connect/systems/activations"
	resp, err := DoGET(urlSuffix)
	if err != nil {
		log.Fatal(err)
	}
	return ParseJSON(resp)
}
