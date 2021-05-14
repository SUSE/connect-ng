package connect

import (
	"encoding/json"
	"gitlab.suse.de/doreilly/go-connect/connect/xlog"
	"time"
)

// Status is used to create JSON output for a product
type Status struct {
	Identifier string `json:"identifier"`
	Version    string `json:"version"`
	Arch       string `json:"arch"`
	Status     string `json:"status"`
}

// NewStatus creates a Status for non-registered Product
func NewStatus(p Product) Status {
	return Status{
		Identifier: p.Name,
		Version:    p.Version,
		Arch:       p.Arch,
		Status:     "Not Registered",
	}
}

// StatusReg adds additional fields to Status for a registered product
type StatusReg struct {
	Status
	RegCode   string    `json:"regcode"`
	StartsAt  time.Time `json:"starts_at"`
	ExpiresAt time.Time `json:"expires_at"`
	SubStatus string    `json:"subscription_status"`
	Type      string    `json:"type"`
}

// NewStatusReg creates a StatusReg from an Activation
func NewStatusReg(a Activation) StatusReg {
	return StatusReg{
		Status: Status{
			Identifier: a.Service.Product.Identifier,
			Version:    a.Service.Product.Version,
			Arch:       a.Service.Product.Arch,
			Status:     "Registered"},
		RegCode:   a.RegCode,
		StartsAt:  a.StartsAt,
		ExpiresAt: a.ExpiresAt,
		SubStatus: a.Status,
		Type:      a.Type,
	}
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

func getStatuses() []interface{} {
	products := GetInstalledProducts()
	activations := GetActivations()

	activationMap := make(map[string]Activation)
	for _, activation := range activations {
		activationMap[activation.ToTriplet()] = activation
	}

	var statuses []interface{}
	for _, product := range products {
		key := product.ToTriplet()
		activation, inMap := activationMap[key]
		if inMap && !activation.IsFree() {
			statuses = append(statuses, NewStatusReg(activation))
		} else {
			statuses = append(statuses, NewStatus(product))
		}
	}
	return statuses
}
