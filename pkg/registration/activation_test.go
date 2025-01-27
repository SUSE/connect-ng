package registration_test

import (
	"testing"

	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFetchProductActivations(t *testing.T) {
	assert := assert.New(t)

	conn, creds := mockConnectionWithCredentials()
	login, password, _ := creds.Login()

	payload := fixture(t, "pkg/registration/activations.json")
	conn.On("Do", mock.Anything).Return(payload, nil).Run(checkAuthBySystemCredentials(t, login, password))

	activations, err := registration.FetchActivations(conn)

	assert.NoError(err)
	assert.Equal(4, len(activations))

	sles := &registration.Activation{}
	for _, activation := range activations {
		if activation.RegistrationCode == "SOME_TEST_REGCODE" {
			sles = activation
		}
	}
	assert.Equal("15.6", sles.Product.Version)
	assert.Equal("SUSE_Linux_Enterprise_Server_15_SP6_x86_64", sles.Metadata.Name)
}

func TestFetchProductActivationsEmpty(t *testing.T) {
	assert := assert.New(t)

	conn, creds := mockConnectionWithCredentials()
	login, password, _ := creds.Login()

	conn.On("Do", mock.Anything).Return([]byte("[]"), nil).Run(checkAuthBySystemCredentials(t, login, password))

	activations, err := registration.FetchActivations(conn)

	assert.NoError(err)
	assert.Equal(0, len(activations))
}
