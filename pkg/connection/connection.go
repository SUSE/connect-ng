package connection

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	DefaultAPIVersion = "application/json,application/vnd.scc.suse.com.v4+json"
)

// Connection is to be implemented by any struct that attempts to perform
// requests against a remote resource that implements the /connect API.
type Connection interface {
	// Returns an HTTP request object that can be used by a subsequent `Do`
	// call.
	BuildRequest(verb string, path string, body any) (*http.Request, error)

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

func (conn ApiConnection) BuildRequest(verb string, path string, body any) (*http.Request, error) {
	bodyData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(verb, conn.Options.URL, bytes.NewReader(bodyData))
	if err != nil {
		return nil, err
	}
	request.URL.Path = path

	conn.setupGenericHeaders(request)

	return request, nil
}

func (conn ApiConnection) Do(request *http.Request) ([]byte, error) {
	token, tokenErr := conn.Credentials.Token()
	if tokenErr != nil {
		return []byte{}, tokenErr
	}
	request.Header.Set("System-Token", token)

	response, doErr := conn.setupHTTPClient().Do(request)
	if doErr != nil {
		return nil, doErr
	}
	defer response.Body.Close()

	// Update the credentials from the new system token.
	token = response.Header.Get("System-Token")
	if err := conn.Credentials.UpdateToken(token); err != nil {
		return nil, err
	}

	// Check if there was an error from the given API response.
	if apiError := ErrorFromResponse(response); apiError != nil {
		return nil, apiError
	}

	data, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return nil, readErr
	}
	return data, nil
}

func (conn ApiConnection) GetCredentials() Credentials {
	return conn.Credentials
}

func (conn ApiConnection) setupGenericHeaders(request *http.Request) {
	userAgent := fmt.Sprintf("%s/%s", conn.Options.AppName, conn.Options.Version)

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", DefaultAPIVersion)
	request.Header.Set("User-Agent", userAgent)

	if conn.Options.PreferedLanguage != "" {
		request.Header.Set("Accept-Language", conn.Options.PreferedLanguage)
	}
}

func (conn ApiConnection) setupHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: conn.Options.Secure}

	if conn.Options.Proxy != nil {
		transport.Proxy = conn.Options.Proxy
	}

	return &http.Client{Transport: transport, Timeout: conn.Options.Timeout}
}
