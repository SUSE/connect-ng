package connect

import (
	"encoding/json"
	"net/http"
)

const (
	// APIVersion is the SCC API version
	APIVersion = "v4"
)

// announceSystem announces a system to SCC
// https://scc.suse.com/connect/v4/documentation#/subscriptions/post_subscriptions_systems
// The body parameter is produced by makeSysInfoBody()
func announceSystem(body []byte) (string, string, error) {
	resp, err := callHTTP("POST", "/connect/subscriptions/systems", body, nil, authToken)
	if err != nil {
		return "", "", err
	}
	var creds struct {
		Login     string `json:"login"`
		Passoword string `json:"password"`
	}
	if err = json.Unmarshal(resp, &creds); err != nil {
		return "", "", err
	}
	return creds.Login, creds.Passoword, nil
}

func upToDate() bool {
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

// systemActivations returns a map keyed by "Identifier/Version/Arch"
func systemActivations() (map[string]Activation, error) {
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

func showProduct(productQuery Product) (Product, error) {
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
