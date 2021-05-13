package connect

import (
	"io"
	"net/http"
)

func DoGET(urlSuffix string) ([]byte, error) {
	config := LoadConfig()
	credentials, err := GetCredentials()
	if err != nil {
		return nil, err
	}
	url := config.BaseURL + urlSuffix
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(credentials.Username, credentials.Password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//log.Printf("resp: %s", string(body))
	return body, nil
}
