package registration

import (
	"errors"
	"testing"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	hostname = "test-hostname"
)

var noProfileData = map[string]any{}

func TestStatusRegistered(t *testing.T) {
	assert := assert.New(t)

	_, conn, creds := connection.NewMockConnectionWithCredentials()
	login, password, _ := creds.Login()

	// 204 No Content
	conn.On("Do", mock.Anything).Return([]byte(""), nil).Run(checkAuthBySystemCredentials(t, login, password))

	status, err := Status(conn, hostname, NoSystemInformation, noProfileData, NoExtraData)
	assert.NoError(err)
	assert.Equal(Registered, status)
}

func TestStatusUknown(t *testing.T) {
	assert := assert.New(t)

	_, conn, _ := connection.NewMockConnectionWithCredentials()

	// 404 Not Found
	conn.On("Do", mock.Anything).Return([]byte(""), errors.New("system not found"))

	status, err := Status(conn, hostname, NoSystemInformation, noProfileData, NoExtraData)
	assert.ErrorContains(err, "system not found")
	assert.Equal(Unknown, status)
}

func TestStatusWithSystemInformation(t *testing.T) {
	assert := assert.New(t)

	systemInformation := map[string]any{
		"key": "value",
	}

	_, conn, _ := connection.NewMockConnectionWithCredentials()

	// 204 No Content
	expected := string(fixture(t, "pkg/registration/status_with_system_information.json"))
	conn.On("Do", mock.AnythingOfType("*http.Request")).Return([]byte(""), nil).Run(matchBody(t, expected))

	status, err := Status(conn, hostname, systemInformation, noProfileData, NoExtraData)
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

	_, conn, _ := connection.NewMockConnectionWithCredentials()

	// 204 No Content
	expected := string(fixture(t, "pkg/registration/status_with_extra_data.json"))
	conn.On("Do", mock.AnythingOfType("*http.Request")).Return([]byte(""), nil).Run(matchBody(t, expected))

	status, err := Status(conn, hostname, NoSystemInformation, noProfileData, extraData)
	assert.NoError(err)
	assert.Equal(Registered, status)
}

func TestStatusWithProfileData(t *testing.T) {
	assert := assert.New(t)

	profileData := map[string]any{
		"pci_data:": []string{
			"identifier:9c90e32ca4d6f7c106d98069539bdbef9078c349f03b474031ad702854c03f1c",
			"data:somedata",
		},
	}

	_, conn, _ := connection.NewMockConnectionWithCredentials()

	// 204 No Content
	expected := string(fixture(t, "pkg/registration/status_with_profile_data.json"))
	conn.On("Do", mock.AnythingOfType("*http.Request")).Return([]byte(""), nil).Run(matchBody(t, expected))

	status, err := Status(conn, hostname, NoSystemInformation, profileData, NoExtraData)
	assert.NoError(err)
	assert.Equal(Registered, status)
}
