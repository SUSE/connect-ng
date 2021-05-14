package connect

import (
	"encoding/json"
	"gitlab.suse.de/doreilly/go-connect/connect/xlog"
	"time"
)

// Status is used to create JSON output
type Status struct {
	Identifier string     `json:"identifier"`
	Version    string     `json:"version"`
	Arch       string     `json:"arch"`
	Status     string     `json:"status"`
	RegCode    string     `json:"regcode,omitempty"`
	StartsAt   *time.Time `json:"starts_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	SubStatus  string     `json:"subscription_status,omitempty"`
	Type       string     `json:"type,omitempty"`
}

// GetProductStatuses returns statuses of installed products
func GetProductStatuses(format string) string {
	if format == "json" {
		statuses := getStatuses()
		jsonStr, err := json.Marshal(statuses)
		if err != nil {
			xlog.Error.Fatal(err)
		}
		return string(jsonStr)
	}

	if format == "text" {
		statuses := getStatuses()
		xlog.Debug.Printf("+%v\n", statuses)
		return "Not implamented \n"
	}
	panic("Parameter must be \"json\" or \"text\"")
}

func getStatuses() []Status {
	products := GetInstalledProducts()
	activations := GetActivations()

	activationMap := make(map[string]Activation)
	for _, activation := range activations {
		activationMap[activation.ToTriplet()] = activation
	}

	var statuses []Status
	for _, product := range products {
		status := Status{
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
			status.StartsAt = &activation.StartsAt
			status.ExpiresAt = &activation.ExpiresAt
			status.SubStatus = activation.Status
			status.Type = activation.Type
			status.Status = "Registered"
		}
		statuses = append(statuses, status)
	}
	return statuses
}
