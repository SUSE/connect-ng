package connect

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

const (
	defaulCredentialsDir  = "/etc/zypp/credentials.d"
	globalCredentialsFile = "/etc/zypp/credentials.d/SCCcredentials"
)

var (
	userMatch = regexp.MustCompile(`(?m)^\s*username\s*=\s*(\S+)\s*$`)
	passMatch = regexp.MustCompile(`(?m)^\s*password\s*=\s*(\S+)\s*$`)
)

// Credentials stores the SCC or service credentials
type Credentials struct {
	Filename string
	Username string
	Password string
}

func (c Credentials) String() string {
	return fmt.Sprintf("file: %s, username: %s, password: REDACTED",
		c.Filename, c.Username)
}

func systemCredentialsFile() string {
	return filepath.Join(CFG.FsRoot, globalCredentialsFile)
}

func serviceCredentialsFile(service string) string {
	return filepath.Join(CFG.FsRoot, defaulCredentialsDir, service)
}

// getCredentials reads the system credentials from the SCCcredentials file
func getCredentials() (Credentials, error) {
	path := systemCredentialsFile()
	if !fileExists(path) {
		return Credentials{}, ErrMissingCredentialsFile
	}
	return readCredentials(path)
}

func readCredentials(path string) (Credentials, error) {
	Debug.Print("Reading credentials: ", path)
	f, err := os.Open(path)
	if err != nil {
		return Credentials{}, err
	}
	defer f.Close()
	creds, err := parseCredentials(f)
	if err != nil {
		return Credentials{}, err
	}
	creds.Filename = path
	Debug.Print("Credentials read: ", creds)
	return creds, nil
}

func parseCredentials(r io.Reader) (Credentials, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return Credentials{}, err
	}
	uMatch := userMatch.FindStringSubmatch(string(content))
	pMatch := passMatch.FindStringSubmatch(string(content))
	if len(uMatch) != 2 || len(pMatch) != 2 {
		return Credentials{}, ErrMalformedSccCredFile
	}
	return Credentials{Username: uMatch[1], Password: pMatch[1]}, nil
}

func (c Credentials) write() error {
	Debug.Print("Writing credentials: ", c)
	dir := filepath.Dir(c.Filename)
	if !fileExists(dir) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	buf := bytes.Buffer{}
	fmt.Fprintln(&buf, "username=", c.Username)
	fmt.Fprintln(&buf, "password=", c.Password)
	return os.WriteFile(c.Filename, buf.Bytes(), 0600)
}

func writeSystemCredentials(login, password string) error {
	path := systemCredentialsFile()
	c := Credentials{Filename: path, Username: login, Password: password}
	return c.write()
}

func writeServiceCredentials(serviceName string) error {
	c, err := getCredentials()
	if err != nil {
		return err
	}
	path := serviceCredentialsFile(serviceName)
	s := Credentials{
		Filename: path,
		Username: c.Username,
		Password: c.Password}
	return s.write()
}

func removeSystemCredentials() error {
	path := systemCredentialsFile()
	return removeFile(path)
}

func removeServiceCredentials(serviceName string) error {
	Debug.Print("Removing service credentials for: ", serviceName)
	path := serviceCredentialsFile(serviceName)
	return removeFile(path)
}
