package connection

import (
	"bytes"
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	DefaultAPIVersion = "application/json,application/vnd.scc.suse.com.v4+json"
)

var (
	//go:embed version.txt
	rawVersion string
)

// Connection is to be implemented by any struct that attempts to perform
// requests against a remote resource that implements the /connect API.
type Connection interface {
	// Returns an HTTP request object that can be used by a subsequent `Do`
	// call.
	BuildRequest(verb string, path string, body any) (*http.Request, error)

	// Performs an HTTP request to the remote API. Returns the status code, response body or
	// an error object.
	Do(*http.Request) (int, []byte, error)

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

func (conn ApiConnection) Do(request *http.Request) (int, []byte, error) {
	token, tokenErr := conn.Credentials.Token()
	if tokenErr != nil {
		return 0, []byte{}, tokenErr
	}
	request.Header.Add("System-Token", token)

	response, doErr := conn.setupHTTPClient().Do(request)
	if doErr != nil {
		return 0, nil, doErr
	}
	defer response.Body.Close()
	code := response.StatusCode

	// Update the credentials from the new system token.
	token = response.Header.Get("System-Token")
	if err := conn.Credentials.UpdateToken(token); err != nil {
		return code, nil, err
	}

	if !successCode(response.StatusCode) {
		msg := parseError(response.Body)
		return code, nil, fmt.Errorf("API error: %v (code: %v)", msg, response.StatusCode)
	}

	data, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return code, nil, readErr
	}
	return code, data, nil
}

func (conn ApiConnection) GetCredentials() Credentials {
	return conn.Credentials
}

func (conn ApiConnection) setupGenericHeaders(request *http.Request) {
	version := strings.Split(strings.TrimSpace(rawVersion), "~")[0]
	userAgent := fmt.Sprintf("%s/%s/%s", version, conn.Options.AppName, conn.Options.Version)

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", DefaultAPIVersion)
	request.Header.Add("User-Agent", userAgent)

	if conn.Options.PreferedLanguage != "" {
		request.Header.Add("Accept-Language", conn.Options.PreferedLanguage)
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

func successCode(code int) bool {
	return code >= 200 && code < 300
}

func parseError(body io.Reader) string {
	var errResp struct {
		Error          string `json:"error"`
		LocalizedError string `json:"localized_error"`
	}
	if err := json.NewDecoder(body).Decode(&errResp); err != nil {
		return ""
	}
	if errResp.LocalizedError != "" {
		return errResp.LocalizedError
	}
	return errResp.Error
}