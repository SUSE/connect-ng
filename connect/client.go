package connect

import (
	"fmt"
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
		return nil, fmt.Errorf("%s", err)
	}
	req.SetBasicAuth(credentials.Username, credentials.Password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	//log.Printf("resp: %s", string(body))
	return body, nil
}
