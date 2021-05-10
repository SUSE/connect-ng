package connect

import (
	"log"
)

func GetActivations() []Activation {
	urlSuffix := "connect/systems/activations"
	resp, err := DoGET(urlSuffix)
	if err != nil {
		log.Println(err)
		return []Activation{}
	}
	return ParseJSON(resp)
}
