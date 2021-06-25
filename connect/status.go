package connect

import (
	"bytes"
	_ "embed" //golint
	"encoding/json"
	"text/template"
)

var (
	//go:embed status-text.tmpl
	statusTemplate string
)

// Status is used to create JSON output
type Status struct {
	Summary    string `json:"-"`
	Identifier string `json:"identifier"`
	Version    string `json:"version"`
	Arch       string `json:"arch"`
	Status     string `json:"status"`
	RegCode    string `json:"regcode,omitempty"`
	StartsAt   string `json:"starts_at,omitempty"`
	ExpiresAt  string `json:"expires_at,omitempty"`
	SubStatus  string `json:"subscription_status,omitempty"`
	Type       string `json:"type,omitempty"`
}

// GetProductStatuses returns statuses of installed products
func GetProductStatuses(format string) (string, error) {
	statuses, err := getStatuses()
	if err != nil {
		return "", err
	}
	if format == "json" {
		statusList := make([]Status, 0, len(statuses))
		for _, s := range statuses {
			statusList = append(statusList, s)
		}
		jsonStr, err := json.Marshal(statusList)
		if err != nil {
			return "", err
		}
		return string(jsonStr), nil
	}

	text, err := getStatusText(statuses)
	if err != nil {
		return "", err
	}
	return text, nil
}

func getStatuses() (map[string]Status, error) {
	statuses := make(map[string]Status)
	products, err := installedProducts()
	if err != nil {
		return statuses, err
	}

	activations := make(map[string]Activation) // default empty map
	if IsRegistered() {
		activations, err = GetActivations()
		if err != nil {
			return statuses, err
		}
	}

	for _, product := range products {
		status := Status{
			Summary:    product.Summary,
			Identifier: product.Name,
			Version:    product.Version,
			Arch:       product.Arch,
			Status:     "Not Registered",
		}
		if activation, ok := activations[product.toTriplet()]; ok {
			status.Status = "Registered"
			if !activation.isFree() {
				status.RegCode = activation.RegCode
				layout := "2006-01-02 15:04:05 MST"
				status.StartsAt = activation.StartsAt.Format(layout)
				status.ExpiresAt = activation.ExpiresAt.Format(layout)
				status.SubStatus = activation.Status
				status.Type = activation.Type
			}
		}
		statuses[product.toTriplet()] = status
	}
	return statuses, nil
}

func getStatusText(statuses map[string]Status) (string, error) {
	t, err := template.New("status-text").Parse(statusTemplate)
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	err = t.Execute(&output, statuses)
	if err != nil {
		return "", err
	}
	return output.String(), nil
}
