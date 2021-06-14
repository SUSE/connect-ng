package connect

// Product represents an installed product
type Product struct {
	Name    string `xml:"name,attr"`
	Version string `xml:"version,attr"`
	Arch    string `xml:"arch,attr"`
	Summary string `xml:"summary,attr"`
	IsBase  bool   `xml:"isbase,attr"`
}

func (p Product) toTriplet() string {
	return p.Name + "/" + p.Version + "/" + p.Arch
}
