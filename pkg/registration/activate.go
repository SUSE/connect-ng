package registration

import "github.com/SUSE/connect-ng/pkg/connection"

// Metadata holds all the data that is returned by activate/deactivate API calls
// which is not exactly tied to the Product struct. Note that by pairing a
// filled Metadata object and a Product object could give you, for example, a
// Zypper service.
type Metadata struct {
	// ID of the activation as given by SCC's API.
	ID int `json:"id"`

	// URL of the product activation so it can be used by other clients (e.g.
	// zypper).
	URL string `json:"url"`

	// Name of the product activation.
	Name string `json:"name"`

	// Extra name that is provided by SCC's APIs.
	ObsoletedName string `json:"obsoleted_service_name"`
}

// Activate a product by pairing an authorized connection (which contains the
// system at hand), plus the "triplet" being used to identify the desired
// product.
func Activate(conn connection.Connection, identifier, version, arch string) (Metadata, Product, error) {
	return Metadata{}, Product{}, nil
}

// Deactivate a product by pairing an authorized connection (which contains the
// system at hand), plus the "triplet" being used to identify the product to be
// deactivated for the system.
func Deactivate(conn connection.Connection, identifier, version, arch string) (Metadata, Product, error) {
	return Metadata{}, Product{}, nil
}
