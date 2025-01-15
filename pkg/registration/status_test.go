package registration_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/SUSE/connect-ng/pkg/registration"
)

const (
	hostname = "test-hostname"
)

func TestStatusRegistered(t *testing.T) {
	assert := assert.New(t)

	conn, creds := mockConnectionWithCredentials()
	login, password, _ := creds.Login()

	// 204 No Content
	conn.On("Do", mock.Anything).Return([]byte(""), nil).Run(checkAuthBySystemCredentials(t, login, password))

	status, err := registration.Status(conn, hostname, nil)
	assert.NoError(err)
	assert.Equal(registration.Registered, status)
}

func TestStatusUnregistered(t *testing.T) {
	assert := assert.New(t)

	conn, _ := mockConnectionWithCredentials()

	// 404 Not Found
	conn.On("Do", mock.Anything).Return([]byte(""), errors.New("system not found"))

	status, err := registration.Status(conn, hostname, nil)
	assert.NoError(err)
	assert.Equal(registration.Unregistered, status)
}

func TestStatusWithSystemInformation(t *testing.T) {
	assert := assert.New(t)

	payload := map[string]any{
		"key": "value",
	}

	conn, _ := mockConnectionWithCredentials()

	// 204 No Content
	expected := string(fixture(t, "pkg/registration/status_with_system_information.json"))
	conn.On("Do", mock.AnythingOfType("*http.Request")).Return([]byte(""), nil).Run(matchBody(t, expected))

	status, err := registration.Status(conn, hostname, payload)
	assert.NoError(err)
	assert.Equal(registration.Registered, status)
}
