package connect

import (
	"strings"
)

// Product represents an installed product or product information from API
type Product struct {
	Name    string `xml:"name,attr" json:"identifier"`
	Version string `xml:"version,attr" json:"version"`
	Arch    string `xml:"arch,attr" json:"arch"`
	Summary string `xml:"summary,attr"`
	IsBase  bool   `xml:"isbase,attr"`

	FriendlyName string `json:"friendly_name"`
	Available    bool   `json:"available"`
	Free         bool   `json:"free"`
	// optional extension products
	Extensions []Product `json:"extensions"`
}

func (p Product) toTriplet() string {
	return p.Name + "/" + p.Version + "/" + p.Arch
}

func (p Product) toQuery() map[string]string {
	return map[string]string{
		"identifier": p.Name,
		"version":    p.Version,
		"arch":       p.Arch,
	}
}

func (p Product) distroTarget() string {
	identifier := strings.ToLower(p.Name)
	if strings.HasPrefix(identifier, "sle") {
		identifier = "sle"
	}
	version := strings.Split(p.Version, ".")[0]
	return identifier + "-" + version + "-" + p.Arch
}
