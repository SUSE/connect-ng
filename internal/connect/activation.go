package connect

import (
	"time"
)

// Activation mimics the shape of the json from the api
type Activation struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	RegCode   string    `json:"regcode"`
	Type      string    `json:"type"`
	StartsAt  time.Time `json:"starts_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Service   Service   `json:"service"`
}

func (a Activation) toTriplet() string {
	p := a.Service.Product
	return p.Name + "/" + p.Version + "/" + p.Arch
}
