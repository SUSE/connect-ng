package connect

import (
	"io"
	"log"
	"net/http"
)

func DoGET(urlSuffix string) ([]byte, error) {
	config := LoadConfig()
	credentials, err := GetCredentials()
	if err != nil {
		log.Fatal(err)
	}
	url := config.BaseURL + urlSuffix
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(credentials.Username, credentials.Password)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	//log.Printf("resp: %s", string(body))
	return body, nil
}
