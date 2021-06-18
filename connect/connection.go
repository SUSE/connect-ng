package connect

import (
	"crypto/tls"
	"encoding/json"
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

// DoGET performs http client GET calls
func DoGET(creds Credentials, urlSuffix string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: CFG.Insecure},
		Proxy:           http.ProxyFromEnvironment,
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", CFG.BaseURL, nil)
	if err != nil {
		return nil, err
	}
	req.URL.Path = urlSuffix
	req.SetBasicAuth(creds.Username, creds.Password)
	addHeaders(req)
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
		return nil, APIError{Message: errMsg, Code: resp.StatusCode}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
