package connect

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	cred "github.com/SUSE/connect-ng/internal/credentials"
	"github.com/SUSE/connect-ng/internal/util"
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

	// Pass the current system token.
	creds, err := cred.ReadCredentials(cred.SystemCredentialsPath(CFG.FsRoot))
	token := ""
	if err == nil {
		token = creds.SystemToken
	}
	req.Header.Add("System-Token", token)
}

func addAuthHeader(req *http.Request, auth authType) error {
	switch auth {
	case authSystem:
		c, err := cred.ReadCredentials(cred.SystemCredentialsPath(CFG.FsRoot))
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

// Returns true if proxy setup is enabled at the system level. This is specific
// to SUSE.
func proxyEnabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("PROXY_ENABLED")))

	// NOTE: if the value is not set, we return true so Go figures this out.
	return value == "" || value == "y" || value == "yes" || value == "t" || value == "true"
}

func proxyWithAuth(req *http.Request) (*url.URL, error) {
	// Check for the special "PROXY_ENABLED" environment variable which might be
	// set in a SUSE system. If it is set to a falsey value, then we skip proxy
	// detection regardless of other environment variables.
	if !proxyEnabled() {
		return nil, nil
	}

	proxyURL, err := http.ProxyFromEnvironment(req)
	// nil proxyUrl indicates no proxy configured
	if proxyURL == nil || err != nil {
		return proxyURL, err
	}
	// add or replace proxy credentials if configured
	if c, err := cred.ReadCurlrcCredentials(); err == nil {
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

	if util.IsLoggerEnabled(util.Debug) {
		reqBlob, _ := httputil.DumpRequestOut(req, true)
		util.Debug.Printf("%s\n", reqBlob)
	}

	resp, err := httpclient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// If we failed to detect which server type was being used when loading the
	// configuration, we can actually further inspect it via some of the headers
	// that are returned by Glue vs RMT. Hence, if the server type is unknown,
	// make an educated guess now.
	if CFG.ServerType == UnknownProvider {
		if api := resp.Header.Get("Scc-Api-Version"); api == sccAPIVersion {
			CFG.ServerType = SccProvider
		} else {
			CFG.ServerType = RmtProvider
		}
	}

	// For each request SCC might update the System token for a given system.
	// This will be given through the `System-Token` header, so we have to grab
	// this here and store it for the next request.
	if err := cred.HandleSystemToken(resp.Header.Get("System-Token"), CFG.FsRoot); err != nil {
		util.Debug.Printf("system-token: %s\n", err)
	}

	if util.IsLoggerEnabled(util.Debug) {
		respBlob, _ := httputil.DumpResponse(resp, true)
		util.Debug.Printf("%s\n", respBlob)
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

// ReloadCertPool triggers reload of internals CA cert pool
func ReloadCertPool() error {
	// TODO: update when https://github.com/golang/go/issues/41888 is fixed
	httpclient = nil
	setupHTTPClient()
	return nil
}

func downloadFile(url string) ([]byte, error) {
	setupHTTPClient()

	resp, err := httpclient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if !successCode(resp.StatusCode) {
		return nil, fmt.Errorf("Downloading %s failed (code: %d): %s", url, resp.StatusCode, resBody)
	}
	return resBody, nil
}
