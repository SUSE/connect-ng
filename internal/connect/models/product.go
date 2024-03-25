package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/SUSE/connect-ng/internal/zypper"
)

// Product represents an installed product or product information from API
//
// NOTE (FTW epic): some of the things here do not map correctly with SCC's API
// and it's admittedly quite bananas (e.g. SCC's "identifier" being "Name" but
// then SCC's "name" being "LongName" and claiming that it's used by Yast
// (wtf?)).
type Product struct {
	Name    string `xml:"name,attr" json:"identifier"`
	Version string `xml:"version,attr" json:"version"`
	Arch    string `xml:"arch,attr" json:"arch"`
	Release string `xml:"release,attr" json:"-"`
	Summary string `xml:"summary,attr" json:"summary,omitempty"`
	IsBase  bool   `xml:"isbase,attr" json:"isbase"`

	FriendlyName string `json:"friendly_name,omitempty"`
	ReleaseType  string `xml:"registerrelease,attr" json:"release_type,omitempty"`
	ProductLine  string `xml:"productline,attr"`
	Available    bool   `json:"available"`
	Free         bool   `json:"free"`
	Recommended  bool   `json:"recommended"`
	// optional extension products
	Extensions []Product `json:"extensions,omitempty"`

	// these are used by YaST
	ID           int    `json:"-"` // handled by custom unmarshaller/marshaller
	Description  string `xml:"description" json:"description,omitempty"`
	EULAURL      string `json:"eula_url,omitempty"`
	FormerName   string `json:"former_identifier,omitempty"`
	ProductType  string `json:"product_type,omitempty"`
	ShortName    string `json:"shortname,omitempty"`
	LongName     string `json:"name,omitempty"`
	ReleaseStage string `json:"release_stage,omitempty"`
	Repositories []Repository `json:"repositories,omitempty"`
}

// UnmarshalJSON custom unmarshaller for Product.
// Special decoding is needed for the Available, IsBase and ID fields.
func (p *Product) UnmarshalJSON(data []byte) error {
	// Only SMT/RMT send the "available" field in their JSON responses.
	// SCC does not, and the default Unmarshal() sets Available to the
	// boolean zero-value which is false. This sets it to true instead.
	type product Product
	prod := product{
		Available: true,
	}
	if err := json.Unmarshal(data, &prod); err != nil {
		return err
	}

	// migration paths contain is-base information as "base" attribute
	// while we default to "isbase" for YaST integration.
	// the rest of the SCC API uses `product_type="base"` instead.
	mProd := struct {
		Base bool `json:"base"`
	}{}
	if err := json.Unmarshal(data, &mProd); err != nil {
		return err
	}
	prod.IsBase = prod.IsBase || mProd.Base || prod.ProductType == "base"

	// SCC and RMT servers send the ID field as an int. But SMT sends it
	// as a string in the POST migrations call. This helper tries both.
	prod.ID = findID(data)

	*p = Product(prod)
	return nil
}

// findID decodes the "id" field from data
func findID(data []byte) int {
	// Try to decode as an int first. SCC/RMT case.
	idInt := struct {
		ID int `json:"id"`
	}{}
	if err := json.Unmarshal(data, &idInt); err == nil {
		return idInt.ID
	}

	// Decoding id as an int failed. Try to decode as a string. SMT case.
	idStr := struct {
		ID int `json:"id,string"`
	}{}
	if err := json.Unmarshal(data, &idStr); err == nil {
		return idStr.ID
	}

	return 0
}

// MarshalJSON is a custom JSON marshaller that includes the "id" field.
// This method is needed because the `json:"id"` tag can not be used
// on Product.ID because the that field requires a custom unmarshaller.
func (p *Product) MarshalJSON() ([]byte, error) {
	type prodAlias Product
	return json.Marshal(&struct {
		ID int `json:"id"`
		*prodAlias
	}{
		ID:        p.ID,
		prodAlias: (*prodAlias)(p),
	})
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

func (p Product) findExtension(query Product) (Product, error) {
	for _, e := range p.Extensions {
		if e.ToTriplet() == query.ToTriplet() {
			return e, nil
		}
		if len(e.Extensions) > 0 {
			if child, err := e.findExtension(query); err == nil {
				return child, nil
			}
		}
	}
	return Product{}, fmt.Errorf("Extension not found")
}

// parseProductsXML returns products parsed from zypper XML
func parseProductsXML(xmlDoc []byte) ([]Product, error) {
	var products struct {
		Products []Product `xml:"product-list>product"`
	}
	if err := xml.Unmarshal(xmlDoc, &products); err != nil {
		return []Product{}, err
	}
	// override ProductType with OEM value if defined
	for i, p := range products.Products {
		if oemValue, err := oemReleaseType(p.ProductLine); err == nil {
			products.Products[i].ReleaseType = oemValue
		}
	}
	return products.Products, nil
}

// installedProducts returns installed products
func InstalledProducts() ([]Product, error) {
    xml, err := zypper.GetInstalledProductsXML()
	if err != nil {
		return []Product{}, err
	}
	return parseProductsXML(xml)
}


// from package_search.go
// SearchPackageProduct represents product reference in package search result
type SearchPackageProduct struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Ident   string `json:"identifier"`
	Type    string `json:"type"`
	Free    bool   `json:"free"`
	Edition string `json:"edition"`
	Arch    string `json:"architecture"`
}

// from client.go
type ProductOut struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	Version    string `json:"version"`
	Arch       string `json:"arch"`
}
