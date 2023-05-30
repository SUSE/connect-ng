package connect

import (
	"bytes"
	_ "embed" //golint
	"encoding/json"
	"text/template"
)

const (
	registered    = "Registered"
	notRegistered = "Not Registered"
)

var (
	//go:embed status-text.tmpl
	statusTemplate string
	//go:embed system-info-text.tmpl
	systemInfoTemplate string
)

// Status is used to create the JSON for `SUSEConnect --status`.
// And to render the template for `SUSEConnect --status-text`.
//
// TODO(josegomezr): Improve naming of this struct.
//
//	it reflects Product Status alongisde Subscription status.
type Status struct {
	Name       string `json:"name,omitempty"`
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

func GetSystemInformation(format string) (string, error) {
	// we're going to explicitly ignore the error to see how much of the partial
	// structure we got to fill.

	// TODO(josegomezr): SendKeepAlivePing sets these two vars as empty. we need
	//                   better design here.
	sysinfo, _ := MakeSystemInformation("", "", nil)
	if format == "json" {
		jsn, err := json.Marshal(sysinfo)
		if err != nil {
			return "", err
		}
		return string(jsn), nil
	}

	text, err := renderSysInfoText(sysinfo)
	if err != nil {
		return "", err
	}
	return text, nil
}

// GetProductStatuses returns statuses of installed products
func GetProductStatuses(format string) (string, error) {
	statuses, err := getStatuses()
	if err != nil {
		return "", err
	}
	if format == "json" {
		jsn, err := json.Marshal(statuses)
		if err != nil {
			return "", err
		}
		return string(jsn), nil
	}

	text, err := getStatusText(statuses)
	if err != nil {
		return "", err
	}
	return text, nil
}

func getStatuses() ([]Status, error) {
	products, err := installedProducts()
	if err != nil {
		return nil, err
	}

	activations := make(map[string]Activation) // default empty map
	if IsRegistered() {
		activations, err = systemActivations()
		if err != nil {
			return nil, err
		}
	}
	statuses := buildStatuses(products, activations)
	return statuses, nil
}

func buildStatuses(products []Product, activations map[string]Activation) []Status {
	var statuses []Status
	for _, product := range products {
		status := Status{
			Summary:    product.Summary,
			Identifier: product.Name,
			Version:    product.Version,
			Arch:       product.Arch,
			Status:     notRegistered,
		}
		if activation, ok := activations[product.ToTriplet()]; ok {
			status.Status = registered
			if activation.RegCode != "" {
				status.Name = activation.Name
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
	return statuses
}

func renderSysInfoText(sysinfo SystemInformation) (string, error) {
	var tpl = template.Must(template.New("system-info-text").Parse(systemInfoTemplate))
	var output bytes.Buffer
	if err := tpl.Execute(&output, sysinfo); err != nil {
		return "", err
	}
	return output.String(), nil
}

func getStatusText(statuses []Status) (string, error) {
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

// SystemProducts returns sum of installed and activated products
// Products from zypper have priority over products from
// activations as they have summary field which is missing
// in the latter.
func SystemProducts() ([]Product, error) {
	products, err := installedProducts()
	if err != nil {
		return products, err
	}
	installedIDs := NewStringSet()
	for _, prod := range products {
		installedIDs.Add(prod.ToTriplet())
	}
	if !IsRegistered() {
		return products, nil
	}
	activations, err := systemActivations()
	if err != nil {
		return products, err
	}
	for _, a := range activations {
		if !installedIDs.Contains(a.Service.Product.ToTriplet()) {
			products = append(products, a.Service.Product)
		}
	}

	return products, nil
}
