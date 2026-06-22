package main

import (
	"fmt"

	"github.com/google/uuid"
)

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
	randomToken, _ := uuid.NewRandom()
	fmt.Printf("<- randomized token: %s\n", randomToken)

	return randomToken.String(), nil
}

func (creds *SccCredentials) UpdateToken(token string) error {
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
