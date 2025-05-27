package registration

import (
	"errors"
	"testing"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterSuccess(t *testing.T) {
	assert := assert.New(t)

	conn, creds := connection.NewMockConnectionWithCredentials()

	// 204 No Content
	response := fixture(t, "pkg/registration/announce_success.json")

	conn.On("Do", mock.Anything).Return(response, nil).Run(checkAuthByRegcode(t, "regcode"))
	creds.On("SetLogin", "SCC_login", "sample-password").Return(nil)

	_, err := Register(conn, "regcode", "hostname", NoSystemInformation, NoExtraData)
	assert.NoError(err)

	conn.AssertExpectations(t)
}

func TestRegisterFailed(t *testing.T) {
	assert := assert.New(t)

	conn, _ := connection.NewMockConnectionWithCredentials()

	// 404 Not found / announce failed
	conn.On("Do", mock.Anything).Return([]byte{}, errors.New("Invalid registration token supplied"))

	_, err := Register(conn, "regcode", "hostname", NoSystemInformation, NoExtraData)
	assert.ErrorContains(err, "Invalid registration token")

	conn.AssertExpectations(t)
}

func TestRegsiterWithSystemInformation(t *testing.T) {
	assert := assert.New(t)

	systemInformation := map[string]any{
		"cpus":    3,
		"sockets": 3,
		"sap": []map[string]any{
			{
				"system_id":      "DEV",
				"instance_types": []string{"ASCS"},
			},
		}}

	response := fixture(t, "pkg/registration/announce_success.json")
	body := fixture(t, "pkg/registration/register_with_system_information.json")

	conn, creds := connection.NewMockConnectionWithCredentials()
	conn.On("Do", mock.Anything).Return(response, nil).Run(matchBody(t, string(body)))
	creds.On("SetLogin", "SCC_login", "sample-password").Return(nil)

	_, err := Register(conn, "regcode", "hostname", systemInformation, NoExtraData)
	assert.NoError(err)

	conn.AssertExpectations(t)
}

func TestRegisterWithEnrichedAttributes(t *testing.T) {
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

	response := fixture(t, "pkg/registration/announce_success.json")
	body := fixture(t, "pkg/registration/register_with_extra_data.json")

	conn, creds := connection.NewMockConnectionWithCredentials()
	conn.On("Do", mock.Anything).Return(response, nil).Run(matchBody(t, string(body)))
	creds.On("SetLogin", "SCC_login", "sample-password").Return(nil)

	_, err := Register(conn, "regcode", "hostname", NoSystemInformation, extraData)
	assert.NoError(err)

	conn.AssertExpectations(t)
}

func TestDeRegisterSuccess(t *testing.T) {
	assert := assert.New(t)

	conn, creds := connection.NewMockConnectionWithCredentials()

	// 404 Not found / announce failed
	conn.On("Do", mock.Anything).Return([]byte{}, nil)
	creds.On("SetLogin", "", "").Return(nil)

	err := Deregister(conn)
	assert.NoError(err)

	conn.AssertExpectations(t)
}

func TestDeRegisterInvalid(t *testing.T) {
	assert := assert.New(t)

	conn, _ := connection.NewMockConnectionWithCredentials()

	// 404 Not found / announce failed
	conn.On("Do", mock.Anything).Return([]byte{}, errors.New("Not found"))

	err := Deregister(conn)
	assert.Error(err)

	conn.AssertExpectations(t)
}
