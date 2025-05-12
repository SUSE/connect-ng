package main

import "fmt"

type SccCredentials struct {
	SystemLogin string
	Password    string
	SystemToken string
	ShowTraces  bool
}

func (SccCredentials) HasAuthentication() bool {
	return true
}

func (creds *SccCredentials) Token() (string, error) {
	if creds.ShowTraces {
		fmt.Printf("<- fetch token %s\n", creds.SystemToken)
	}
	return creds.SystemToken, nil
}

func (creds *SccCredentials) UpdateToken(token string) error {
	if creds.ShowTraces {
		fmt.Printf("-> update token %s\n", token)
	}
	creds.SystemToken = token
	return nil
}

func (creds *SccCredentials) Login() (string, string, error) {
	if creds.SystemLogin == "" || creds.Password == "" {
		return "", "", fmt.Errorf("login credentials not set")
	}

	if creds.ShowTraces {
		fmt.Printf("<- fetch login %s\n", creds.SystemLogin)
	}
	return creds.SystemLogin, creds.Password, nil
}

func (creds *SccCredentials) SetLogin(login, password string) error {
	if creds.ShowTraces {
		fmt.Printf("-> set login %s\n", login)
	}
	creds.SystemLogin = login
	creds.Password = password
	return nil
}
