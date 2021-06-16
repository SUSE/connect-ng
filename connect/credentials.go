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
	errParseCredientials = errors.New("Unable to parse credentials")
	userMatch            = regexp.MustCompile(`(?m)^\s*username\s*=\s*(\S+)\s*$`)
	passMatch            = regexp.MustCompile(`(?m)^\s*password\s*=\s*(\S+)\s*$`)
)

// Credentials stores the SCC credentials
// TODO service credentials
type Credentials struct {
	Username string
	Password string
}

func getCredentials() (Credentials, error) {
	Debug.Printf("Reading credientials from %s", defaulCredPath)
	f, err := os.Open(defaulCredPath)
	if err != nil {
		return Credentials{}, err
	}
	defer f.Close()
	return parseCredientials(f)
}

func parseCredientials(r io.Reader) (Credentials, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return Credentials{}, err
	}
	uMatch := userMatch.FindStringSubmatch(string(content))
	pMatch := passMatch.FindStringSubmatch(string(content))
	if len(uMatch) != 2 || len(pMatch) != 2 {
		return Credentials{}, errParseCredientials
	}
	return Credentials{Username: uMatch[1], Password: pMatch[1]}, nil
}
