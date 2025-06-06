package registration

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/SUSE/connect-ng/pkg/connection"
)

// Product as defined from SCC'S API.
type Product struct {
	// NOTE: zypper does not handle this, hence XML parsing is left out.
	Identifier string `json:"identifier"`

	Name    string `xml:"name,attr" json:"name"`
	Version string `xml:"version,attr" json:"version"`
	Arch    string `xml:"arch,attr" json:"arch"`
	Release string `xml:"release,attr" json:"-"`
	Summary string `xml:"summary,attr" json:"summary,omitempty"`
	IsBase  bool   `xml:"isbase,attr" json:"isbase"`

	FriendlyName string `json:"friendly_name,omitempty"`
	ReleaseType  string `xml:"registerrelease,attr" json:"release_type,omitempty"`
	Available    bool   `json:"available"`
	Free         bool   `json:"free"`
	Recommended  bool   `json:"recommended"`

	// Optional extension products
	Extensions []Product `json:"extensions,omitempty"`

	Description  string       `xml:"description" json:"description,omitempty"`
	EULAURL      string       `json:"eula_url,omitempty"`
	FormerName   string       `json:"former_identifier,omitempty"`
	ProductType  string       `json:"product_type,omitempty"`
	ProductLine  string       `xml:"productline,attr"`
	ShortName    string       `json:"shortname,omitempty"`
	ReleaseStage string       `json:"release_stage,omitempty"`
	Repositories []Repository `json:"repositories,omitempty"`
}

// Builds a new Product object by parsing the given string considering to be a
// product "triplet" (i.e. a string with the format "<name>/<version>/<arch>").
func FromTriplet(triplet string) (Product, error) {
	if match, _ := regexp.MatchString(`^\S+/\S+/\S+$`, triplet); !match {
		return Product{}, fmt.Errorf("invalid product; <internal name>/<version>/<architecture> format expected")
	}

	parts := strings.Split(triplet, "/")
	return Product{Name: parts[0], Version: parts[1], Arch: parts[2]}, nil
}

// ToTriplet returns <name>/<version>/<arch> string for this product.
func (p Product) ToTriplet() string {
	return p.Name + "/" + p.Version + "/" + p.Arch
}

// Returns true if the product is empty, false otherwise.
func (p Product) IsEmpty() bool {
	return p.Name == "" || p.Version == "" || p.Arch == ""
}

// Returns VERSION[-RELEASE] for the current product.
func (p Product) Edition() string {
	if p.Release == "" {
		return p.Version
	}
	return p.Version + "-" + p.Release
}

// Returns a map which transforms the current product to something that can be
// used as a query for an HTTP request.
func (p Product) ToQuery() map[string]string {
	return map[string]string{
		"identifier":   p.Name,
		"version":      p.Version,
		"arch":         p.Arch,
		"release_type": p.ReleaseType,
	}
}

// TraverseFunc is called for each extension of the given product.
// If true is returned, traversal is continued.
type TraverseFunc func(product Product) (bool, error)

// Returns the products triplet identifier.
func (p *Product) ToTriplet() string {
	return p.Identifier + "/" + p.Version + "/" + p.Arch
}

// TraverseExtensions traverse through the products extensions and theirs extensions.
// When TraverseFunc returns false, the full product and its extensions are skipped.
func (pro *Product) TraverseExtensions(fn TraverseFunc) error {
	for _, extension := range pro.Extensions {
		doContinue, fnErr := fn(extension)

		if fnErr != nil {
			return fnErr
		}

		if doContinue {
			if err := extension.TraverseExtensions(fn); err != nil {
				return err
			}
		}
	}
	return nil
}

// Returns the extension for this product which matches the given triplet
// identifier.
func (p Product) FindExtension(triplet string) (Product, error) {
	for _, e := range p.Extensions {
		if e.ToTriplet() == triplet {
			return e, nil
		}
		if len(e.Extensions) > 0 {
			if child, err := e.FindExtension(triplet); err == nil {
				return child, nil
			}
		}
	}
	return Product{}, fmt.Errorf("extension not found")
}

// Transforms the current product into a list of extensions.
func (p Product) ToExtensionsList() []Product {
	res := make([]Product, 0)
	for _, e := range p.Extensions {
		res = append(res, e)
		res = append(res, e.ToExtensionsList()...)
	}
	return res
}

type productShowRequest struct {
	Identifier string `json:"identifier"`
	Version    string `json:"version"`
	Arch       string `json:"arch"`
}

// FetchProductInfo fetches information about a product specified by identifier (e.g. SLES) and its version and architecture.
// The Result also includes the available extension tree, which can be used to activate leaf extensions.
func FetchProductInfo(conn connection.Connection, identifier, version, arch string) (*Product, error) {
	payload := productShowRequest{
		Identifier: identifier,
		Version:    version,
		Arch:       arch,
	}
	creds := conn.GetCredentials()
	login, password, credErr := creds.Login()

	if credErr != nil {
		return nil, credErr
	}

	request, buildErr := conn.BuildRequest("GET", "/connect/systems/products", payload)
	if buildErr != nil {
		return nil, buildErr
	}

	connection.AddSystemAuth(request, login, password)

	response, doErr := conn.Do(request)
	if doErr != nil {
		return nil, doErr
	}

	product := Product{}

	if err := json.Unmarshal(response, &product); err != nil {
		return nil, err
	}

	return &product, nil
}

// Updates the given product (i.e. upgrades/downgrades it) depending on what the
// SCC server determines. An SCC service is given in return.
func UpdateProduct(conn connection.Connection, product Product) (Service, error) {
	return touchProduct(conn, product, "PUT")
}

// Removes the given product as part of the system (i.e. deactivates it). An SCC
// service is given in return.
func RemoveProduct(conn connection.Connection, product Product) (Service, error) {
	return touchProduct(conn, product, "DELETE")
}

// Calls Connect's /connect/system/products API endpoint with the given `verb`
// and for the given `product` as the payload.
func touchProduct(conn connection.Connection, product Product, verb string) (Service, error) {
	creds := conn.GetCredentials()
	login, password, credErr := creds.Login()
	if credErr != nil {
		return Service{}, credErr
	}

	request, buildErr := conn.BuildRequest(verb, "/connect/systems/products", product)
	if buildErr != nil {
		return Service{}, buildErr
	}

	connection.AddSystemAuth(request, login, password)

	response, doErr := conn.Do(request)
	if doErr != nil {
		return Service{}, doErr
	}

	service := Service{}
	if err := json.Unmarshal(response, &service); err != nil {
		return Service{}, err
	}
	return service, nil
}

// Returns true if the given `product` can be found in the list of
// `activations`.
func ProductInActivations(product *Product, activations []*Activation) bool {
	triplet := product.ToTriplet()

	for _, activation := range activations {
		if triplet == activation.Product.ToTriplet() {
			return true
		}
	}
	return false
}
