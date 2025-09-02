package helpers

import (
	"encoding/xml"
	"fmt"
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
	env := NewEnv(zypp.t)
	result, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println(env.RedactRegcodes(string(result)))
		assert.FailNow(zypp.t, fmt.Sprintf("Running zypper failed with exit code %d", cmd.ProcessState.ExitCode()), err)

		return []byte{}
	}

	return result
}

func (zypp *Zypper) FetchProducts() []registration.Product {
	var products struct {
		Products []registration.Product `xml:"product-list>product"`
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
			return product.Identifier, product.Version, product.Arch
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
