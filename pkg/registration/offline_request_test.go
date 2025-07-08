package registration

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBuildOfflineRequest(t *testing.T) {
	assert := assert.New(t)
	uuid := "0f2fb933-00b8-483f-b63a-47d481c12947"

	request := BuildOfflineRequest("rancher", "2.9.4", "x86_64", uuid, NoSystemInformation)

	assert.Equal("rancher", request.Product.Identifier)
	assert.Equal(uuid, request.UUID)
}

func TestOfflineRequestSetCredentials(t *testing.T) {
	assert := assert.New(t)

	uuid := "0f2fb933-00b8-483f-b63a-47d481c12947"
	creds := &connection.MockCredentials{}
	creds.On("Login").Return("login", "password", nil)

	request := BuildOfflineRequest("rancher", "2.9.4", "x86_64", uuid, NoSystemInformation)
	request.SetCredentials(creds)

	assert.Equal("login", request.Login)
	assert.Equal("password", request.Password)

	creds.AssertExpectations(t)
}

func TestOfflineRequestSetServerURL(t *testing.T) {
	assert := assert.New(t)

	uuid := "0f2fb933-00b8-483f-b63a-47d481c12947"
	expected := "http://server-url.com"

	request := BuildOfflineRequest("rancher", "2.9.4", "x86_64", uuid, NoSystemInformation)
	request.SetServerURL(expected)
	assert.Equal(expected, request.ServerURL)
}

func TestBuildBase64Encoded(t *testing.T) {
	assert := assert.New(t)

	uuid := "0f2fb933-00b8-483f-b63a-47d481c12947"
	request := BuildOfflineRequest("rancher", "2.9.4", "x86_64", uuid, NoSystemInformation)

	encoded, encodeErr := request.Base64Encoded()
	assert.NoError(encodeErr)

	var result map[string]any
	decoder := base64.NewDecoder(base64.StdEncoding, encoded)

	unmarshalErr := json.NewDecoder(decoder).Decode(&result)
	assert.NoError(unmarshalErr)

	assert.NotNil(result)
	assert.Equal(uuid, result["uuid"])

	productMap, ok := result["product"].(map[string]any)
	assert.True(ok)

	assert.Equal("rancher", productMap["identifier"])
}

func TestRegisterWithOfflineRequest(t *testing.T) {
	assert := assert.New(t)

	uuid := "0f2fb933-00b8-483f-b63a-47d481c12947"
	regcode := "some-scc-regcode"

	request := BuildOfflineRequest("rancher", "2.9.4", "x86_64", uuid, NoSystemInformation)
	conn, _ := connection.NewMockConnectionWithCredentials()

	body := fixture(t, "pkg/registration/offline_request_request.base64")
	response := fixture(t, "pkg/registration/offline_request_response.base64")

	conn.On("Do", mock.Anything).Return(response, nil).Run(func(args mock.Arguments) {
		matchBody(t, string(body))
		checkAuthByRegcode(t, regcode)
	})

	cert, err := RegisterWithOfflineRequest(conn, regcode, request)
	assert.NoError(err)

	matches, regErr := cert.RegcodeMatches(regcode)
	assert.NoError(regErr)
	assert.True(matches)
}
