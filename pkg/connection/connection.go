package connection

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// Connection is to be implemented by any struct that attempts to perform
// requests against a remote resource that implements the /connect API.
type Connection interface {
	// Returns an HTTP request object that can be used by a subsequent `Do`
	// call.
	GetRequest(string, string, any) (*http.Request, error)

	// Performs an HTTP request to the remote API. Returns the response body or
	// an error object.
	Do(*http.Request) ([]byte, error)

	// Returns the credentials object to be used for authenticated requests.
	GetCredentials() Credentials
}

// ApiConnection implements the 'Connection' interface, providing access to any
// server implementing the /connect API (see
// https://scc.suse.com/connect/v4/documentation for more info).
type ApiConnection struct {
	Options     Options
	Credentials Credentials
}

// Returns an ApiConnection object initialized with the given Options and
// Credentials.
func New(opts Options, creds Credentials) *ApiConnection {
	return &ApiConnection{Options: opts, Credentials: creds}
}

func (conn ApiConnection) GetRequest(verb string, path string, body any) (*http.Request, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(verb, conn.Options.Url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.URL.Path = path

	return req, nil
}

func (conn ApiConnection) Do(req *http.Request) ([]byte, error) {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	// TODO: note that we need to get the RootCAs thingie from the old connection.
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: conn.Options.Secure}
	tr.Proxy = conn.Options.Proxy
	httpclient := &http.Client{Transport: tr, Timeout: 60 * time.Second}

	// TODO: add headers (except auth)

	resp, err := httpclient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// TODO: handle system token
	// TODO: handle success/bad code

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return resBody, nil
}

func (conn ApiConnection) GetCredentials() Credentials {
	return conn.Credentials
}
