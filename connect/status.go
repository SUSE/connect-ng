package connect

import (
	_ "embed" //golint
	"encoding/json"
	"fmt"
	"strings"
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
func GetProductStatuses(format string) string {
	statuses, err := getStatuses()
	if err != nil {
		Error.Println(err)
		return fmt.Sprintf("ERROR: %s", err)
	}
	if format == "json" {
		jsonStr, err := json.Marshal(statuses)
		if err != nil {
			Error.Println(err)
			return fmt.Sprintf("ERROR: %s", err)
		}
		return string(jsonStr)
	}
	if format == "text" {
		text, err := getStatusText(statuses)
		if err != nil {
			Error.Println(err)
			return fmt.Sprintf("ERROR: %s", err)
		}
		return text
	}
	return `ERROR: parameter must be "json" or "text"`
}

func getStatuses() ([]Status, error) {
	var statuses []Status
	products, err := installedProducts()
	if err != nil {
		return statuses, err
	}

	activations := make(map[string]Activation) // default empty map
	if CredentialsExists() {
		creds, err := GetCredentials()
		if err != nil {
			return statuses, err
		}
		cfg := LoadConfig(defaultConfigPath)
		activations, err = GetActivations(cfg, creds)
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
		if activation, ok := activations[product.ToTriplet()]; ok {
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
		statuses = append(statuses, status)
	}
	return statuses, nil
}

func getStatusText(statuses []Status) (string, error) {
	t, err := template.New("status-text").Parse(statusTemplate)
	if err != nil {
		return "", err
	}
	var outWriter strings.Builder
	err = t.Execute(&outWriter, statuses)
	if err != nil {
		return "", err
	}
	return outWriter.String(), nil
}
