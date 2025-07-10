package helpers

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/stretchr/testify/assert"
)

// Proxy implements a simple HTTP proxy server which can be setup to assert
// certain conditions on how the SUSEConnect binary interacts with proxies.
type Proxy struct {
	// Assert object from the feature test.
	Assert *assert.Assertions

	// The expected authorization from the `Proxy-Authorization` header
	// (e.g. the value setup in `$HOME/.curlrc`.
	ExpectedProxyAuth string
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	proxyHeader := req.Header.Get("Proxy-Authorization")
	credentials, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(proxyHeader, "Basic "))

	p.Assert.NoError(err)
	p.Assert.Equal(string(credentials), p.ExpectedProxyAuth)

	w.WriteHeader(http.StatusOK)
}
