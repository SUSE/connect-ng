package registration_test

import (
	"errors"
	"testing"

	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterSuccess(t *testing.T) {
	assert := assert.New(t)

	conn, creds := mockConnectionWithCredentials()

	// 204 No Content
	payload := fixture(t, "pkg/registration/announce_success.json")

	conn.On("Do", mock.Anything).Return(204, payload, nil).Run(checkAuthByRegcode(t, "regcode"))
	creds.On("SetLogin", "SCC_login", "sample-password").Return(nil)

	_, err := registration.Register(conn, "regcode", "hostname", nil)
	assert.NoError(err)

	conn.AssertExpectations(t)
}

func TestRegisterFailed(t *testing.T) {
	assert := assert.New(t)

	conn, _ := mockConnectionWithCredentials()

	// 404 Not found / announce failed
	conn.On("Do", mock.Anything).Return(404, []byte{}, errors.New("Invalid registration token supplied"))

	_, err := registration.Register(conn, "regcode", "hostname", nil)
	assert.ErrorContains(err, "Invalid registration token")

	conn.AssertExpectations(t)
}

func TestDeRegisterSuccess(t *testing.T) {
	assert := assert.New(t)

	conn, creds := mockConnectionWithCredentials()

	// 404 Not found / announce failed
	conn.On("Do", mock.Anything).Return(204, []byte{}, nil)
	creds.On("SetLogin", "", "").Return(nil)

	err := registration.Deregister(conn)
	assert.NoError(err)

	conn.AssertExpectations(t)
}

func TestDeRegisterInvalid(t *testing.T) {
	assert := assert.New(t)

	conn, _ := mockConnectionWithCredentials()

	// 404 Not found / announce failed
	conn.On("Do", mock.Anything).Return(404, []byte{}, errors.New("Not found"))

	err := registration.Deregister(conn)
	assert.Error(err)

	conn.AssertExpectations(t)
}
