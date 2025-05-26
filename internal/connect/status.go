package connect

import (
	"bytes"
	_ "embed" //golint
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/SUSE/connect-ng/internal/zypper"
	"github.com/SUSE/connect-ng/pkg/registration"
)

type StatusFormat int

const (
	StatusJSON StatusFormat = iota
	StatusText
)

const (
	registered    = "Registered"
	notRegistered = "Not Registered"
)

var (
	//go:embed status-text.tmpl
	statusTemplate string
)

// Status is used to create the JSON for `SUSEConnect --status`.
// And to render the template for `SUSEConnect --status-text`.
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

func PrintProductStatuses(opts *Options, format StatusFormat) error {
	output, err := GetProductStatuses(opts, format)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

// GetProductStatuses returns statuses of installed products
func GetProductStatuses(opts *Options, format StatusFormat) (string, error) {
	statuses, err := getStatuses(opts)
	if err != nil {
		return "", err
	}

	switch format {
	case StatusJSON:
		jsn, err := json.Marshal(statuses)
		if err != nil {
			return "", err
		}
		return string(jsn), nil
	case StatusText:
		text, err := getStatusText(statuses)
		if err != nil {
			return "", err
		}
		return text, nil
	}
	// Never happens. Hooray for Go's enums and branch exhaustion!
	return "", nil
}

func getStatuses(opts *Options) ([]Status, error) {
	installed, err := zypper.InstalledProducts()
	if err != nil {
		return nil, err
	}

	api := NewWrappedAPI(opts)

	activations := make(map[string]*registration.Activation) // default empty map
	if api.IsRegistered() {
		rawActivations, err := registration.FetchActivations(api.GetConnection())
		if err != nil {
			return nil, err
		}
		for _, activation := range rawActivations {
			activations[activation.ToTriplet()] = activation
		}
	}
	installedProducts := zypperProductListToProductList(installed)
	statuses := buildStatuses(installedProducts, activations)
	return statuses, nil
}

func buildStatuses(products []Product, activations map[string]*registration.Activation) []Status {
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
			if activation.RegistrationCode != "" {
				status.Name = activation.Name
				status.RegCode = activation.RegistrationCode
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
	installed, err := zypper.InstalledProducts()
	products := zypperProductListToProductList(installed)
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
