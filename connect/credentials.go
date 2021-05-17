package connect

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
)

const (
	defaulCredPath = "/etc/zypp/credentials.d/SCCcredentials"
)

var (
	ErrNoCredentialsFile = errors.New("Credentials file does not exist")
	ErrParseCredientials = errors.New("Unable to parse credentials")
	userMatch            = regexp.MustCompile(`(?m)^\s*username\s*=\s*(\S+)\s*$`)
	passMatch            = regexp.MustCompile(`(?m)^\s*password\s*=\s*(\S+)\s*$`)
)

type Credentials struct {
	Username string
	Password string
}

func credentialsEqual(a, b Credentials) bool {
	return a.Username == b.Username && a.Password == b.Password
}

func GetCredentials() (Credentials, error) {
	Debug.Printf("Reading credientials from %s", defaulCredPath)
	f, err := os.Open(defaulCredPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Credentials{}, fmt.Errorf("%w; %s", ErrNoCredentialsFile, err)
		}
		return Credentials{}, err
	}
	defer f.Close()
	return ParseCredientials(f)
}

func ParseCredientials(r io.Reader) (Credentials, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return Credentials{}, err
	}
	uMatch := userMatch.FindStringSubmatch(string(content))
	pMatch := passMatch.FindStringSubmatch(string(content))
	if len(uMatch) != 2 || len(pMatch) != 2 {
		return Credentials{}, ErrParseCredientials
	}
	return Credentials{Username: uMatch[1], Password: pMatch[1]}, nil
}
