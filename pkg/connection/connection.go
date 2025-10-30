package connection

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
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
	// Builds a http.Request and setup up headers. The body can provided as json marshable object
	// The created request can be used in a subsequent `Do` call.
	BuildRequest(verb string, path string, body any) (*http.Request, error)

	// Builds a http.Request and setup up headers. The body can provided as io.Reader
	BuildRequestRaw(verb string, path string, body io.Reader) (*http.Request, error)

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
	var reader io.Reader
	buffer := bytes.Buffer{}

	if body != nil {
		// Make sure we do not encode HTML data which might interfere with instance_data which can be valid
		// XML.
		encoder := json.NewEncoder(&buffer)
		encoder.SetEscapeHTML(false)

		err := encoder.Encode(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(buffer.Bytes())
	}
	return conn.BuildRequestRaw(verb, path, reader)
}

func (conn ApiConnection) BuildRequestRaw(verb string, path string, body io.Reader) (*http.Request, error) {
	fullUrl := fmt.Sprintf("%s%s", conn.Options.URL, path)
	request, err := http.NewRequest(verb, fullUrl, body)
	if err != nil {
		return nil, err
	}

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
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: !conn.Options.Secure}

	if conn.Options.Proxy != nil {
		transport.Proxy = conn.Options.Proxy
	}

	if conn.Options.Certificate != nil {
		pool := x509.NewCertPool()
		pool.AddCert(conn.Options.Certificate)

		transport.TLSClientConfig.RootCAs = pool
	}

	return &http.Client{Transport: transport, Timeout: conn.Options.Timeout}
}
