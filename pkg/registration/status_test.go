package registration

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

	status, err := Status(conn, hostname, NoSystemInformation, NoExtraData)
	assert.NoError(err)
	assert.Equal(Registered, status)
}

func TestStatusUnregistered(t *testing.T) {
	assert := assert.New(t)

	conn, _ := mockConnectionWithCredentials()

	// 404 Not Found
	conn.On("Do", mock.Anything).Return([]byte(""), errors.New("system not found"))

	status, err := Status(conn, hostname, NoSystemInformation, NoExtraData)
	assert.NoError(err)
	assert.Equal(Unregistered, status)
}

func TestStatusWithSystemInformation(t *testing.T) {
	assert := assert.New(t)

	systemInformation := map[string]any{
		"key": "value",
	}

	conn, _ := mockConnectionWithCredentials()

	// 204 No Content
	expected := string(fixture(t, "pkg/registration/status_with_system_information.json"))
	conn.On("Do", mock.AnythingOfType("*http.Request")).Return([]byte(""), nil).Run(matchBody(t, expected))

	status, err := Status(conn, hostname, systemInformation, NoExtraData)
	assert.NoError(err)
	assert.Equal(Registered, status)
}

func TestStatusWithExtraData(t *testing.T) {
	assert := assert.New(t)

	extraData := map[string]any{
		"online_at": []string{
			"12122025:000000000000000000000000",
			"11122025:000000000000000000000000",
			"10122025:000000000000000000000000",
		},
		"namespace":     "staging-sles",
		"instance_data": "<document>{\"instanceId\": \"dummy_instance_data\"}</document>",
	}

	conn, _ := mockConnectionWithCredentials()

	// 204 No Content
	expected := string(fixture(t, "pkg/registration/status_with_extra_data.json"))
	conn.On("Do", mock.AnythingOfType("*http.Request")).Return([]byte(""), nil).Run(matchBody(t, expected))

	status, err := Status(conn, hostname, NoSystemInformation, extraData)
	assert.NoError(err)
	assert.Equal(Registered, status)
}
