package connect

// Product represents an installed product
type Product struct {
	Name    string `xml:"name,attr"`
	Version string `xml:"version,attr"`
	Arch    string `xml:"arch,attr"`
	Summary string `xml:"summary,attr"`
	IsBase  bool   `xml:"isbase,attr"`
}

func (p Product) ToTriplet() string {
	return p.Name + "/" + p.Version + "/" + p.Arch
}

func ProductsEqual(p1, p2 Product) bool {
	return p1.Name == p2.Name && p1.Version == p2.Version && p1.Arch == p2.Arch
}
