package registration

import (
	"errors"
	"testing"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFetchSubscriptionInfoSuccess(t *testing.T) {
	assert := assert.New(t)

	conn, _ := connection.NewMockConnectionWithCredentials()

	payload := fixture(t, "pkg/registration/subscription_info.json")
	conn.On("Do", mock.Anything).Return(payload, nil).Run(
		checkAuthByRegcode(t, "test-regcode"),
	)

	info, err := FetchSubscriptionInfo(conn, "test-regcode")
	assert.NoError(err)
	assert.NotNil(info)
	assert.Equal("full", info.Kind)
	assert.Equal("SUSE Linux Enterprise Server with Rancher, x86_64, 3-Year Subscription", info.Name)
	assert.Equal(50, info.Limit)
	assert.Equal("all", info.Notifications)
	assert.Len(info.ProductClasses, 2)
	assert.Equal("RANCHER-X86", info.ProductClasses[0].Name)
	assert.Equal("SUSE Rancher Manager for x86_64 architecture", info.ProductClasses[0].Description)
	assert.Equal("7261", info.ProductClasses[1].Name)
	assert.Equal("SUSE Linux Enterprise Server (x86_64)", info.ProductClasses[1].Description)
}

func TestFetchSubscriptionInfoAPIError(t *testing.T) {
	assert := assert.New(t)

	conn, _ := connection.NewMockConnectionWithCredentials()
	conn.On("Do", mock.Anything).Return([]byte{}, errors.New("Invalid regcode"))

	_, err := FetchSubscriptionInfo(conn, "invalid-regcode")
	assert.Error(err)
}

func TestFetchSubscriptionInfoMalformedJSON(t *testing.T) {
	assert := assert.New(t)

	conn, _ := connection.NewMockConnectionWithCredentials()
	conn.On("Do", mock.Anything).Return([]byte("{invalid json}"), nil)

	_, err := FetchSubscriptionInfo(conn, "test-regcode")
	assert.Error(err)
}

func TestFetchSubscriptionProductsSuccess(t *testing.T) {
	assert := assert.New(t)

	conn, _ := connection.NewMockConnectionWithCredentials()

	payload := fixture(t, "pkg/registration/subscription_products.json")
	conn.On("Do", mock.Anything).Return(payload, nil).Run(
		checkAuthByRegcode(t, "test-regcode"),
	)

	products, err := FetchSubscriptionProducts(conn, "test-regcode")
	assert.NoError(err)
	assert.Len(products, 3)
	assert.Equal("SUSE_SLES-SP2-migration", products[0].Identifier)
	assert.Equal("SUSE Linux Enterprise Server", products[0].Name)
	assert.Equal("i686", products[0].Arch)
	assert.Len(products[0].Extensions, 1)
	assert.Len(products[0].Repositories, 8)
	assert.Equal("rancher", products[1].Identifier)
	assert.Equal("SUSE Rancher", products[1].Name)
	assert.Equal("rancher-prime", products[2].Identifier)
	assert.Equal("SUSE Rancher Prime", products[2].Name)
}

func TestFetchSubscriptionProductsAPIError(t *testing.T) {
	assert := assert.New(t)

	conn, _ := connection.NewMockConnectionWithCredentials()
	conn.On("Do", mock.Anything).Return([]byte{}, errors.New("Invalid regcode"))

	_, err := FetchSubscriptionProducts(conn, "invalid-regcode")
	assert.Error(err)
}

func TestFetchSubscriptionProductsMalformedJSON(t *testing.T) {
	assert := assert.New(t)

	conn, _ := connection.NewMockConnectionWithCredentials()
	conn.On("Do", mock.Anything).Return([]byte("[{invalid json}]"), nil)

	_, err := FetchSubscriptionProducts(conn, "test-regcode")
	assert.Error(err)
}
