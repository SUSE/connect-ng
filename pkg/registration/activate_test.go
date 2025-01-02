package registration_test

import (
	"errors"
	"testing"

	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestActivateProductSuccess(t *testing.T) {
	assert := assert.New(t)

	conn, _ := mockConnectionWithCredentials()

	// 204 No Content
	payload := fixture(t, "pkg/registration/activate_success.json")
	conn.On("Do", mock.Anything).Return(200, payload, nil)

	metadata, product, err := registration.Activate(conn, "SLES", "12.1", "x86_64", "regcode")
	assert.NoError(err)

	assert.Equal("SUSE_Linux_Enterprise_Server_12_x86_64", metadata.ObsoletedName)
	assert.Equal("SUSE Linux Enterprise Server 12 x86_64", product.FriendlyName)
}

func TestActivateProductInvalidRegcode(t *testing.T) {
	assert := assert.New(t)

	conn, _ := mockConnectionWithCredentials()

	// 204 No Content
	conn.On("Do", mock.Anything).Return(422, []byte{}, errors.New("No valid subscription found"))

	_, _, err := registration.Activate(conn, "SLES", "12.1", "x86_64", "regcode")
	assert.Error(err)
}

func TestDeactivateProductSuccess(t *testing.T) {
	assert := assert.New(t)

	conn, _ := mockConnectionWithCredentials()

	// 204 No Content
	payload := fixture(t, "pkg/registration/deactivate_success.json")
	conn.On("Do", mock.Anything).Return(200, payload, nil)

	metadata, product, err := registration.Deactivate(conn, "SLES", "12.1", "x86_64")
	assert.NoError(err)

	assert.Equal("SUSE_Linux_Enterprise_Server_12_x86_64", metadata.ObsoletedName)
	assert.Equal("SUSE Linux Enterprise Server 12 x86_64", product.FriendlyName)
}

func TestDeactivateProductInvalidProduct(t *testing.T) {
	assert := assert.New(t)

	conn, _ := mockConnectionWithCredentials()

	// 204 No Content
	conn.On("Do", mock.Anything).Return(422, []byte{}, errors.New("Product is a base product and cannot be deactivated"))

	_, _, err := registration.Deactivate(conn, "SLES", "12.1", "x86_64")
	assert.Error(err)
}
