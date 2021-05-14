package connect

import (
	"encoding/json"
	"gitlab.suse.de/doreilly/go-connect/connect/xlog"
	"time"
)

// AProduct is a Product from API
type AProduct struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Arch       string `json:"arch"`
	Identifier string `json:"identifier"`
	Free       bool   `json:"free"`
}

type Service struct {
	Product AProduct `json:"product"`
}

type Activation struct {
	Service   Service   `json:"service"`
	Status    string    `json:"status"`
	RegCode   string    `json:"regcode"`
	Type      string    `json:"type"`
	StartsAt  time.Time `json:"starts_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (p AProduct) ToTriplet() string {
	return p.Identifier + "/" + p.Version + "/" + p.Arch
}

func (a Activation) ToTriplet() string {
	return a.Service.Product.ToTriplet()
}

func (a Activation) IsFree() bool {
	return a.Service.Product.Free
}

func ParseJSON(jsonStr []byte) []Activation {
	var activations []Activation
	err := json.Unmarshal(jsonStr, &activations)
	if err != nil {
		xlog.Error.Fatal(err)
	}
	return activations
}
