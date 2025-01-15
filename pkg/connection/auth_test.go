package connection

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testRequest(t *testing.T) *http.Request {
	assert := assert.New(t)

	opts := DefaultOptions("testApp", "1.0", "en_US")
	creds := NoCredentials{}
	conn := New(opts, creds)

	request, buildErr := conn.BuildRequest("GET", "/test/api", nil)
	assert.NoError(buildErr)

	return request
}

func TestAuthByRegcode(t *testing.T) {
	assert := assert.New(t)
	request := testRequest(t)

	regcode := "test"
	expected := "Token token=test"

	AddRegcodeAuth(request, regcode)
	assert.Equal(expected, request.Header.Get("Authorization"))
}

func TestAuthBySystemCredentials(t *testing.T) {
	assert := assert.New(t)
	request := testRequest(t)

	login := "login"
	password := "password"

	AddSystemAuth(request, login, password)
	assert.Equal("Basic bG9naW46cGFzc3dvcmQ=", request.Header.Get("Authorization"))
}
