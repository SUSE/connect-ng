package registration

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProductToTriplet(t *testing.T) {
	assert := assert.New(t)

	productJson := fixture(t, "pkg/registration/product_tree.json")
	product := Product{}

	assert.NoError(json.Unmarshal(productJson, &product))
	assert.Equal("SLES/15.5/x86_64", product.ToTriplet())
}

func TestProductTraverseExtensionsFull(t *testing.T) {
	assert := assert.New(t)

	productJson := fixture(t, "pkg/registration/product_tree.json")
	product := Product{}

	actual := []string{}
	expected := []string{
		"sle-module-basesystem/15.5/x86_64",
		"sle-module-desktop-applications/15.5/x86_64",
		"sle-module-development-tools/15.5/x86_64",
		"sle-module-NVIDIA-compute/15/x86_64",
		"sle-module-sap-business-one/15.5/x86_64",
		"sle-we/15.5/x86_64",
		"sle-module-server-applications/15.5/x86_64",
		"sle-module-web-scripting/15.5/x86_64",
		"sle-module-legacy/15.5/x86_64",
		"sle-module-public-cloud/15.5/x86_64",
		"sle-ha/15.5/x86_64",
		"sle-module-containers/15.5/x86_64",
		"sle-module-transactional-server/15.5/x86_64",
		"sle-module-live-patching/15.5/x86_64",
		"sle-module-python3/15.5/x86_64",
		"PackageHub/15.5/x86_64",
		"sle-module-certifications/15.5/x86_64",
		"SLES-LTSS/15.5/x86_64",
	}

	assert.NoError(json.Unmarshal(productJson, &product))

	err := product.TraverseExtensions(func(p Product) (bool, error) {
		actual = append(actual, fmt.Sprintf("%s/%s/%s", p.Identifier, p.Version, p.Arch))
		return true, nil
	})
	assert.NoError(err)
	assert.Equal(expected, actual)
}

func TestProductTraverseExtensionsSkipBranch(t *testing.T) {
	assert := assert.New(t)

	productJson := fixture(t, "pkg/registration/product_tree.json")
	product := Product{}

	actual := []string{}
	expected := []string{
		"sle-module-basesystem/15.5/x86_64",
		"sle-module-desktop-applications/15.5/x86_64",
		"sle-module-development-tools/15.5/x86_64",
		"sle-module-NVIDIA-compute/15/x86_64",
		"sle-module-sap-business-one/15.5/x86_64",
		"sle-we/15.5/x86_64",
		"sle-module-containers/15.5/x86_64",
		"sle-module-transactional-server/15.5/x86_64",
		"sle-module-live-patching/15.5/x86_64",
		"sle-module-python3/15.5/x86_64",
		"PackageHub/15.5/x86_64",
		"sle-module-certifications/15.5/x86_64",
		"SLES-LTSS/15.5/x86_64",
	}

	assert.NoError(json.Unmarshal(productJson, &product))

	err := product.TraverseExtensions(func(p Product) (bool, error) {
		triplet := fmt.Sprintf("%s/%s/%s", p.Identifier, p.Version, p.Arch)

		// Skip everything based on server-applications
		if triplet == "sle-module-server-applications/15.5/x86_64" {
			return false, nil
		}

		actual = append(actual, triplet)
		return true, nil
	})
	assert.NoError(err)
	assert.Equal(expected, actual)
}

func TestFetchProductInfo(t *testing.T) {
	assert := assert.New(t)

	_, conn, creds := connection.NewMockConnectionWithCredentials()
	login, password, _ := creds.Login()

	payload := fixture(t, "pkg/registration/product_tree.json")
	conn.On("Do", mock.Anything).Return(payload, nil).Run(checkAuthBySystemCredentials(t, login, password))

	product, err := FetchProductInfo(conn, "SLES", "15.5", "x86_64")
	assert.NoError(err)

	assert.Equal("SUSE Linux Enterprise Server 15 SP5 x86_64", product.FriendlyName)
}

func TestFindExtension(t *testing.T) {
	assert := assert.New(t)

	productJson := fixture(t, "pkg/registration/product_tree.json")
	product := Product{}

	assert.NoError(json.Unmarshal(productJson, &product))

	found := []string{
		// Direct extension from the product
		"sle-module-basesystem/15.5/x86_64",

		// Extension for one of the extensions from the product
		"sle-module-desktop-applications/15.5/x86_64",
	}

	for _, triplet := range found {
		_, err := product.FindExtension(triplet)
		assert.NoError(err)
	}

	_, err := product.FindExtension("sle-module-fakesystem/15.5/x86_64")
	assert.Equal("extension not found", err.Error())
}
