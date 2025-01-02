package connection

import (
	"fmt"
	"net/http"
)

func AuthByRegcode(request *http.Request, regcode string) {
	tokenAuth := fmt.Sprintf("Token token=%s", regcode)

	request.Header.Add("Authorization", tokenAuth)
}

func AuthBySystemCredentials(request *http.Request, login string, password string) {
	request.SetBasicAuth(login, password)
}
