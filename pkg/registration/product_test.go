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

func TestProductDistroTargetSle(t *testing.T) {
	assert := assert.New(t)

	productJson := fixture(t, "pkg/registration/product_tree.json")
	product := Product{}

	assert.NoError(json.Unmarshal(productJson, &product))
	assert.Equal("sle-15-x86_64", product.DistroTarget())
}

func TestProductDistroTargetNotSle(t *testing.T) {
	assert := assert.New(t)

	productJson := fixture(t, "pkg/registration/product_tree.json")
	product := Product{}

	assert.NoError(json.Unmarshal(productJson, &product))
	product.Identifier = "not-suse-really"
	assert.Equal("not-suse-really-15-x86_64", product.DistroTarget())
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

	conn, creds := connection.NewMockConnectionWithCredentials()
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

func TestFromTriplet(t *testing.T) {
	assert := assert.New(t)

	t.Run("valid triplets with component validation", func(t *testing.T) {
		tests := []struct {
			triplet      string
			expectedName string
			expectedVer  string
			expectedArch string
		}{
			{
				triplet:      "SLES/15/x86_64",
				expectedName: "SLES",
				expectedVer:  "15",
				expectedArch: "x86_64",
			},
			{
				triplet:      "sle-module-basesystem/15.5/aarch64",
				expectedName: "sle-module-basesystem",
				expectedVer:  "15.5",
				expectedArch: "aarch64",
			},
			{
				triplet:      "sle-module-python3/15.6/ppc64le",
				expectedName: "sle-module-python3",
				expectedVer:  "15.6",
				expectedArch: "ppc64le",
			},
			{
				triplet:      "sle-module-server-applications/15.4/s390x",
				expectedName: "sle-module-server-applications",
				expectedVer:  "15.4",
				expectedArch: "s390x",
			},
		}

		for _, tt := range tests {
			t.Run(tt.triplet, func(t *testing.T) {
				product, err := FromTriplet(tt.triplet)
				assert.NoError(err, "triplet should be valid: %q", tt.triplet)
				assert.NotNil(product)
				assert.Equal(tt.expectedName, product.Identifier, "triplet identifier not expected value")
				assert.Equal(tt.expectedVer, product.Version, "triplet version  not expected value")
				assert.Equal(tt.expectedArch, product.Arch, "triplet arch not expected value")
			})
		}
	})

	t.Run("invalid triplets", func(t *testing.T) {
		tests := []struct {
			triplet string
			name    string
		}{
			{triplet: "", name: "empty string"},
			{triplet: "single", name: "single part"},
			{triplet: "two/parts", name: "two parts"},
			{triplet: "four/parts/too/many", name: "four parts"},
			{triplet: "empty//part", name: "empty part"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := FromTriplet(tt.triplet)
				assert.Error(err, "expected error for: %q (%s)", tt.triplet, tt.name)
			})
		}
	})

}
