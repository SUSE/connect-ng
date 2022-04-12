package connect

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const (
	appName       = "SUSEConnect-ng"
	sccAPIVersion = "v4"
)

type authType int

const (
	authNone authType = iota
	authSystem
	authToken
)

var (
	httpclient *http.Client
)

// parseError returns the error message from a SCC error response
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

func addHeaders(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	accept := "application/json,application/vnd.scc.suse.com." + sccAPIVersion + "+json"
	req.Header.Add("Accept", accept)
	if CFG.Language != "" {
		req.Header.Add("Accept-Language", CFG.Language)
	}
	// REVISIT "Accept-Encoding" - disable gzip commpression on debug?
	req.Header.Add("User-Agent", appName+"/"+GetShortenedVersion())
}

func addAuthHeader(req *http.Request, auth authType) error {
	switch auth {
	case authSystem:
		c, err := getCredentials()
		if err != nil {
			return err
		}
		req.SetBasicAuth(c.Username, c.Password)
	case authToken:
		req.Header.Add("Authorization", "Token token="+CFG.Token)
	}
	return nil
}

func successCode(code int) bool {
	return code >= 200 && code < 300
}

func proxyWithAuth(req *http.Request) (*url.URL, error) {
	proxyURL, err := http.ProxyFromEnvironment(req)
	// nil proxyUrl indicates no proxy configured
	if proxyURL == nil || err != nil {
		return proxyURL, err
	}
	// add or replace proxy credentials if configured
	if c, err := readCurlrcCredentials(curlrcCredentialsFile()); err == nil {
		proxyURL.User = url.UserPassword(c.Username, c.Password)
	}
	return proxyURL, nil
}

func setupHTTPClient() {
	if httpclient == nil {
		// use defaults from DefaultTransport
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: CFG.Insecure, RootCAs: systemRootsPool()}
		tr.Proxy = proxyWithAuth
		httpclient = &http.Client{Transport: tr, Timeout: 60 * time.Second}
	}
}

func callHTTP(verb, path string, body []byte, query map[string]string, auth authType) ([]byte, error) {
	setupHTTPClient()
	if httpclient == nil {
		// use defaults from DefaultTransport
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: CFG.Insecure}
		tr.Proxy = proxyWithAuth
		httpclient = &http.Client{Transport: tr, Timeout: 60 * time.Second}
	}
	req, err := http.NewRequest(verb, CFG.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.URL.Path = path

	if err := addAuthHeader(req, auth); err != nil {
		return nil, err
	}
	addHeaders(req)

	values := req.URL.Query()
	for n, v := range query {
		values.Add(n, v)
	}
	req.URL.RawQuery = values.Encode()

	if isLoggerEnabled(Debug) {
		reqBlob, _ := httputil.DumpRequestOut(req, true)
		Debug.Printf("%s\n", reqBlob)
	}

	resp, err := httpclient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if isLoggerEnabled(Debug) {
		respBlob, _ := httputil.DumpResponse(resp, true)
		Debug.Printf("%s\n", respBlob)
	}

	if !successCode(resp.StatusCode) {
		errMsg := parseError(resp.Body)
		return nil, APIError{Message: errMsg, Code: resp.StatusCode}
	}
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return resBody, nil
}

func ReloadCertPool() error {
	// TODO: update when https://github.com/golang/go/issues/41888 is fixed
	httpclient = nil
	setupHTTPClient()
	return nil
}
