package main

import (
	"bytes"
	"fmt"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/SUSE/connect-ng/pkg/validation"
)

type SccCredentials struct {
	Login       string `json:"login"`
	Password    string `json:"password"`
	SystemToken string `json:"system_token"`
}

func (SccCredentials) HasAuthentication() bool {
	return true
}

func (creds *SccCredentials) Triplet() (string, string, string, error) {
	return creds.Login, creds.Password, creds.SystemToken, nil
}

func (creds *SccCredentials) Load() error {
	creds = SccCredentials{
		Login:       "foo",
		Password:    "bar",
		SystemToken: "",
	}
	return nil
}

func (creds *SccCredentials) Update(login, password, token string) error {
	creds = SccCredentials{
		Login:       login,
		Password:    password,
		SystemToken: token,
	}
	return nil
}

func main() {
	fmt.Println("I'm here")

	opts := connection.SCCOptions()

	// No authentication
	//_ = connection.New(opts, connection.NoCredentials{})

	// With authentication
	conn := connection.New(opts, &SccCredentials{})

	_, _ = registration.Status(conn)
	_, _, _ = validation.OfflineActivation(bytes.NewReader([]byte{}))
}
