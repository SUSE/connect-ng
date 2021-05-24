package connect

import (
	"time"
)

// Activation mimics the shape of the json from the api
type Activation struct {
	Status    string    `json:"status"`
	RegCode   string    `json:"regcode"`
	Type      string    `json:"type"`
	StartsAt  time.Time `json:"starts_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Service   struct {
		Product struct {
			Name       string `json:"name"`
			Version    string `json:"version"`
			Arch       string `json:"arch"`
			Identifier string `json:"identifier"`
			Free       bool   `json:"free"`
		} `json:"product"`
	} `json:"service"`
}

func (a Activation) ToTriplet() string {
	p := a.Service.Product
	return p.Identifier + "/" + p.Version + "/" + p.Arch
}

func (a Activation) IsFree() bool {
	return a.Service.Product.Free
}
