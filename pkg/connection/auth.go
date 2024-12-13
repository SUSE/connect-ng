package connection

import (
	"fmt"
	"net/http"
)

func AddRegcodeAuth(request *http.Request, regcode string) {
	tokenAuth := fmt.Sprintf("Token token=%s", regcode)

	request.Header.Set("Authorization", tokenAuth)
}

func AddSystemAuth(request *http.Request, login string, password string) {
	request.SetBasicAuth(login, password)
}
