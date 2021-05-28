package connect

import (
	"fmt"
	"io"
	"net/http"
)

func DoGET(config Config, creds Credentials, urlSuffix string) ([]byte, error) {
	url := config.BaseURL + "/" + urlSuffix
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(creds.Username, creds.Password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Server response: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//log.Printf("resp: %s", string(body))
	return body, nil
}
