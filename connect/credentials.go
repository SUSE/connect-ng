package connect

import (
	"errors"
	"io"
	"os"
	"regexp"
)

const (
	defaulCredPath = "/etc/zypp/credentials.d/SCCcredentials"
)

var (
	ErrParseCredientials = errors.New("Unable to parse credentials")
	userMatch            = regexp.MustCompile(`(?m)^\s*username\s*=\s*(\S+)\s*$`)
	passMatch            = regexp.MustCompile(`(?m)^\s*password\s*=\s*(\S+)\s*$`)
)

type Credentials struct {
	Username string
	Password string
}

func CredentialsExists() bool {
	if _, err := os.Stat(defaulCredPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func GetCredentials() (Credentials, error) {
	Debug.Printf("Reading credientials from %s", defaulCredPath)
	f, err := os.Open(defaulCredPath)
	if err != nil {
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
