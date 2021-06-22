package connect

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
)

type authType int

const (
	authNone authType = iota
	authSystem
	authToken
)

// parseError returns the error message from a SCC error response
func parseError(body io.Reader) string {
	var m map[string]interface{}
	dec := json.NewDecoder(body)
	if err := dec.Decode(&m); err == nil {
		if errMsg, ok := m["error"].(string); ok {
			return errMsg
		}
	}
	return ""
}

func addHeaders(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	accept := "application/json,application/vnd.scc.suse.com." + APIVersion + "+json"
	req.Header.Add("Accept", accept)
	if CFG.Language != "" {
		req.Header.Add("Accept-Language", CFG.Language)
	}
	// REVISIT "Accept-Encoding" - disable gzip commpression on debug?
	req.Header.Add("User-Agent", AppName+"/"+Version)
	// REVISIT Close - unlike Ruby, Go does not close by default
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

func callHTTP(verb, path string, body io.Reader, query map[string]string, auth authType) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: CFG.Insecure},
		Proxy:           http.ProxyFromEnvironment,
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(verb, CFG.BaseURL, body)
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

	reqBlob, _ := httputil.DumpRequestOut(req, true)
	Debug.Printf("%s\n", reqBlob)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBlob, _ := httputil.DumpResponse(resp, true)
	Debug.Printf("%s\n", respBlob)

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
