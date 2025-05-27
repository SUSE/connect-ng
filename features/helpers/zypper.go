package helpers

import (
	"encoding/xml"
	"os/exec"
	"testing"

	"github.com/SUSE/connect-ng/internal/zypper"
	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/stretchr/testify/assert"
)

type Zypper struct {
	t *testing.T
}

func NewZypper(t *testing.T) *Zypper {
	return &Zypper{
		t: t,
	}
}

func (zypp *Zypper) run(params ...string) []byte {
	cmd := exec.Command("zypper", params...)
	result, err := cmd.CombinedOutput()

	if err != nil {
		assert.FailNow(zypp.t, "Running zypper failed", err)
		return []byte{}
	}

	return result
}

func (zypp *Zypper) FetchProducts() []zypper.ZypperProduct {
	var products struct {
		Products []zypper.ZypperProduct `xml:"product-list>product"`
	}

	doc := zypp.run("--no-refresh", "--non-interactive", "--quiet", "--xmlout", "products", "--installed-only")
	if parseErr := xml.Unmarshal(doc, &products); parseErr != nil {
		assert.FailNow(zypp.t, "Could not parse zypper products", parseErr)
	}

	return products.Products
}

func (zypp *Zypper) FetchServices() []zypper.ZypperService {
	var services struct {
		Services []zypper.ZypperService `xml:"service-list>service"`
	}

	doc := zypp.run("--no-refresh", "--non-interactive", "--quiet", "--xmlout", "services")
	if parseErr := xml.Unmarshal(doc, &services); parseErr != nil {
		assert.FailNow(zypp.t, "Could not parse zypper services", parseErr)
	}

	return services.Services
}

func (zypp *Zypper) BaseProduct() (string, string, string) {
	products := zypp.FetchProducts()

	for _, product := range products {
		if product.IsBase {
			return product.Name, product.Version, product.Arch
		}
	}

	assert.FailNow(zypp.t, "Unable to detect base product")
	return "", "", ""
}

func (zypp *Zypper) ServiceExists(product registration.Product) bool {
	services := zypp.FetchServices()
	serviceName := FriendlyNameToServiceName(product.FriendlyName)

	for _, service := range services {
		if service.Name == serviceName {
			return true
		}
	}
	return false
}
