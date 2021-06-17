package connect

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
