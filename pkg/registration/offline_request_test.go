package registration

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBuildOfflineRequest(t *testing.T) {
	assert := assert.New(t)

	request := BuildOfflineRequest("rancher", "2.9.4", "x86_64", NoSystemInformation)

	assert.Equal("rancher", request.Product.Identifier)
}

func TestOfflineRequestSetCredentials(t *testing.T) {
	assert := assert.New(t)

	creds := &mockCredentials{}
	creds.On("Login").Return("login", "password", nil)

	request := BuildOfflineRequest("rancher", "2.9.4", "x86_64", NoSystemInformation)
	request.SetCredentials(creds)

	assert.Equal("login", request.Login)
	assert.Equal("password", request.Password)

	creds.AssertExpectations(t)
}

func TestBuildBase64Encoded(t *testing.T) {
	assert := assert.New(t)

	request := BuildOfflineRequest("rancher", "2.9.4", "x86_64", NoSystemInformation)

	encoded, encodeErr := request.Base64Encoded()
	assert.NoError(encodeErr)

	var result map[string]any
	decoder := base64.NewDecoder(base64.StdEncoding, encoded)

	unmarshalErr := json.NewDecoder(decoder).Decode(&result)
	assert.NoError(unmarshalErr)

	assert.NotNil(result)

	productMap, ok := result["product"].(map[string]any)
	assert.True(ok)

	assert.Equal("rancher", productMap["identifier"])
}

func TestRegisterWithOfflineRequest(t *testing.T) {
	assert := assert.New(t)

	regcode := "some-scc-regcode"

	request := BuildOfflineRequest("rancher", "2.9.4", "x86_64", NoSystemInformation)
	conn, _ := mockConnectionWithCredentials()

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
