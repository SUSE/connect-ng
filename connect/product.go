package connect

import (
	"strings"
)

// Product represents an installed product or product information from API
type Product struct {
	Name    string `xml:"name,attr" json:"identifier"`
	Version string `xml:"version,attr" json:"version"`
	Arch    string `xml:"arch,attr" json:"arch"`
	Summary string `xml:"summary,attr" json:"-"`
	IsBase  bool   `xml:"isbase,attr" json:"-"`

	FriendlyName string `json:"friendly_name,omitempty"`
	ReleaseType  string `json:"release_type,omitempty"`
	Available    bool   `json:"available"`
	Free         bool   `json:"free"`
	// optional extension products
	Extensions []Product `json:"extensions,omitempty"`
}

func (p Product) isEmpty() bool {
	return p.Name == "" || p.Version == "" || p.Arch == ""
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

func (p Product) toExtensionsList() []Product {
	res := make([]Product, 0)
	for _, e := range p.Extensions {
		res = append(res, e)
		res = append(res, e.toExtensionsList()...)
	}
	return res
}

func (p Product) distroTarget() string {
	identifier := strings.ToLower(p.Name)
	if strings.HasPrefix(identifier, "sle") {
		identifier = "sle"
	}
	version := strings.Split(p.Version, ".")[0]
	return identifier + "-" + version + "-" + p.Arch
}
