package connect

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
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

// DoGET performs http client GET calls
func DoGET(config Config, creds Credentials, urlSuffix string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: config.Insecure},
		Proxy:           http.ProxyFromEnvironment,
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", config.BaseURL, nil)
	if err != nil {
		return nil, err
	}
	req.URL.Path = urlSuffix
	req.SetBasicAuth(creds.Username, creds.Password)
	reqBlob, _ := httputil.DumpRequestOut(req, true)
	Debug.Printf("%s\n", reqBlob)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBlob, _ := httputil.DumpResponse(resp, true)
	Debug.Printf("%s\n", respBlob)

	if resp.StatusCode != http.StatusOK {
		errMsg := parseError(resp.Body)
		return nil, fmt.Errorf("%s: %s", resp.Status, errMsg)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
