package registration

import (
	"encoding/json"

	"github.com/SUSE/connect-ng/pkg/connection"
)

// Product as defined from SCC'S API.
type Product struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	Version    string `json:"version"`
	Arch       string `json:"arch"`
	Summary    string `json:"summary,omitempty"`
	IsBase     bool   `json:"isbase"`

	FriendlyName string `json:"friendly_name,omitempty"`
	ReleaseType  string `json:"release_type,omitempty"`
	Available    bool   `json:"available"`
	Free         bool   `json:"free"`
	Recommended  bool   `json:"recommended"`

	// optional extension products
	Extensions []Product `json:"extensions,omitempty"`

	Description  string       `json:"description,omitempty"`
	EULAURL      string       `json:"eula_url,omitempty"`
	FormerName   string       `json:"former_identifier,omitempty"`
	ProductType  string       `json:"product_type,omitempty"`
	ShortName    string       `json:"shortname,omitempty"`
	ReleaseStage string       `json:"release_stage,omitempty"`
	Repositories []Repository `json:"repositories,omitempty"`
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
