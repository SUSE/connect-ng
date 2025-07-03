package registration

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterSuccess(t *testing.T) {
	assert := assert.New(t)

	conn, creds := mockConnectionWithCredentials()

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

	conn, _ := mockConnectionWithCredentials()

	// 404 Not found / announce failed
	conn.On("Do", mock.Anything).Return([]byte{}, errors.New("Invalid registration token supplied"))

	_, err := Register(conn, "regcode", "hostname", NoSystemInformation, NoExtraData)
	assert.ErrorContains(err, "Invalid registration token")

	conn.AssertExpectations(t)
}

func TestRegisterErrorCases(t *testing.T) {
	assert := assert.New(t)

	type testCase struct {
		mockResponse announceResponse
		err          error
	}

	testCases := []testCase{
		{
			mockResponse: announceResponse{
				Id:       RegistrationSystemIdEmpty,
				Login:    "SCC_login",
				Password: "sample-password",
			},
			err: ErrRegistrationSystemIdEmpty,
		},
		{
			mockResponse: announceResponse{
				Id:       RegistrationSystemIdError,
				Login:    "SCC_login",
				Password: "sample-password",
			},
			err: ErrRegistrationSystemIdError,
		},
		{
			mockResponse: announceResponse{
				Id:       RegistrationSystemIdKeepAlive,
				Login:    "SCC_login",
				Password: "sample-password",
			},
			err: ErrRegistrationSystemIdKeepAlive,
		},
		{
			mockResponse: announceResponse{
				Id:       RegistrationSystemIdOffline,
				Login:    "SCC_login",
				Password: "sample-password",
			},
			err: ErrRegistrationSystemIdOffline,
		},
	}

	for _, tc := range testCases {
		conn, creds := mockConnectionWithCredentials()
		creds.On("SetLogin", "SCC_login", "sample-password").Return(nil)
		data, err := json.Marshal(tc.mockResponse)
		assert.NoError(err)
		conn.On("Do", mock.Anything).Return(data, nil)

		_, regErr := Register(conn, "regcode", "hostname", NoSystemInformation, NoExtraData)
		assert.ErrorIs(regErr, tc.err)
		conn.AssertExpectations(t)
	}

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

	conn, creds := mockConnectionWithCredentials()
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

	conn, creds := mockConnectionWithCredentials()
	conn.On("Do", mock.Anything).Return(response, nil).Run(matchBody(t, string(body)))
	creds.On("SetLogin", "SCC_login", "sample-password").Return(nil)

	_, err := Register(conn, "regcode", "hostname", NoSystemInformation, extraData)
	assert.NoError(err)

	conn.AssertExpectations(t)
}

func TestDeRegisterSuccess(t *testing.T) {
	assert := assert.New(t)

	conn, creds := mockConnectionWithCredentials()

	// 404 Not found / announce failed
	conn.On("Do", mock.Anything).Return([]byte{}, nil)
	creds.On("SetLogin", "", "").Return(nil)

	err := Deregister(conn)
	assert.NoError(err)

	conn.AssertExpectations(t)
}

func TestDeRegisterInvalid(t *testing.T) {
	assert := assert.New(t)

	conn, _ := mockConnectionWithCredentials()

	// 404 Not found / announce failed
	conn.On("Do", mock.Anything).Return([]byte{}, errors.New("Not found"))

	err := Deregister(conn)
	assert.Error(err)

	conn.AssertExpectations(t)
}
