package connection_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func NewTestServerSetupWith(t *testing.T, verb, path string, setup func(http.ResponseWriter)) *httptest.Server {
	assert := assert.New(t)

	handler := func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(verb, request.Method, "HTTP method matches")
		assert.Equal(path, request.URL.Path, "URL path matches")

		setup(response)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	return server
}
