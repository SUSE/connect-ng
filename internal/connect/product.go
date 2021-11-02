package connect

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Product represents an installed product or product information from API
type Product struct {
	Name    string `xml:"name,attr" json:"identifier"`
	Version string `xml:"version,attr" json:"version"`
	Arch    string `xml:"arch,attr" json:"arch"`
	Release string `xml:"release,attr" json:"-"`
	Summary string `xml:"summary,attr" json:"-"`
	IsBase  bool   `xml:"isbase,attr" json:"base"`

	FriendlyName string `json:"friendly_name,omitempty"`
	ReleaseType  string `xml:"registerrelease,attr" json:"release_type,omitempty"`
	ProductLine  string `xml:"productline,attr"`
	Available    bool   `json:"available"`
	Free         bool   `json:"free"`
	Recommended  bool   `json:"recommended"`
	// optional extension products
	Extensions []Product `json:"extensions,omitempty"`

	// these are used by YaST
	ID           int    `json:"id"`
	Description  string `xml:"description" json:"description,omitempty"`
	EULAURL      string `json:"eula_url,omitempty"`
	FormerName   string `json:"former_identifier,omitempty"`
	ProductType  string `json:"product_type,omitempty"`
	LongName     string `json:"name,omitempty"`
	ReleaseStage string `json:"release_stage,omitempty"`
	Repositories []Repo `json:"repositories,omitempty"`
}

// UnmarshalJSON custom unmarshaller for Product.
// Only SMT/RMT send the "available" field in their JSON responses.
// SCC does not, and the default Unmarshal() sets Available to the
// boolean zero-value which is false. This sets it to true instead.
func (p *Product) UnmarshalJSON(data []byte) error {
	type product Product // use type alias to prevent infinite recursion
	prod := product{
		Available: true,
	}
	if err := json.Unmarshal(data, &prod); err != nil {
		return err
	}
	*p = Product(prod)
	return nil
}

// Edition returns VERSION[-RELEASE] for product
func (p Product) Edition() string {
	if p.Release == "" {
		return p.Version
	}
	return p.Version + "-" + p.Release
}

func (p Product) isEmpty() bool {
	return p.Name == "" || p.Version == "" || p.Arch == ""
}

// ToTriplet returns <name>/<version>/<arch> string for product
func (p Product) ToTriplet() string {
	return p.Name + "/" + p.Version + "/" + p.Arch
}

// SplitTriplet returns a product from given or error for invalid input
func SplitTriplet(p string) (Product, error) {
	if match, _ := regexp.MatchString(`^\S+/\S+/\S+$`, p); !match {
		return Product{}, fmt.Errorf("invalid product; <internal name>/<version>/<architecture> format expected")
	}
	parts := strings.Split(p, "/")
	return Product{Name: parts[0], Version: parts[1], Arch: parts[2]}, nil
}

func (p Product) toQuery() map[string]string {
	return map[string]string{
		"identifier":   p.Name,
		"version":      p.Version,
		"arch":         p.Arch,
		"release_type": p.ReleaseType,
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
