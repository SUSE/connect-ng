package connect

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
)

// DoGET performs http client GET calls
func DoGET(config Config, creds Credentials, urlSuffix string) ([]byte, error) {
	client := &http.Client{}
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
		return nil, fmt.Errorf("Server response: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
