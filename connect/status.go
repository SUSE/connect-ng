package connect

import (
	_ "embed" //golint
	"encoding/json"
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
	statuses := getStatuses()
	if format == "json" {
		jsonStr, err := json.Marshal(statuses)
		if err != nil {
			Error.Fatal(err)
		}
		return string(jsonStr)
	}
	if format == "text" {
		return doStatusText(statuses)
	}
	panic("Parameter must be \"json\" or \"text\"")
}

func getStatuses() []Status {
	products := GetInstalledProducts()

	var err error
	activations := []Activation{}
	if CredentialsExists() {
		activations, err = GetActivations()
		if err != nil {
			Error.Fatal(err)
		}
	}

	activationMap := make(map[string]Activation)
	for _, activation := range activations {
		activationMap[activation.ToTriplet()] = activation
	}

	var statuses []Status
	for _, product := range products {
		status := Status{
			Summary:    product.Summary,
			Identifier: product.Name,
			Version:    product.Version,
			Arch:       product.Arch,
			Status:     "Not Registered",
		}
		key := product.ToTriplet()
		activation, inMap := activationMap[key]
		// TODO registered but not activated?
		if inMap && !activation.IsFree() {
			status.RegCode = activation.RegCode
			layout := "2006-01-02 15:04:05 MST"
			status.StartsAt = activation.StartsAt.Format(layout)
			status.ExpiresAt = activation.ExpiresAt.Format(layout)
			status.SubStatus = activation.Status
			status.Type = activation.Type
			status.Status = "Registered"
		}
		statuses = append(statuses, status)
	}
	return statuses
}

func doStatusText(statuses []Status) string {
	t, err := template.New("status-text").Parse(statusTemplate)
	if err != nil {
		Error.Fatal(err)
	}
	var outWriter strings.Builder
	err = t.Execute(&outWriter, statuses)
	if err != nil {
		Error.Fatal(err)
	}
	return outWriter.String()
}
