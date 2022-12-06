package connect

import (
	"bufio"
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
	curlrcUserFile        = ".curlrc"
)

var (
	userMatch        = regexp.MustCompile(`(?m)^\s*username\s*=\s*(\S+)\s*$`)
	passMatch        = regexp.MustCompile(`(?m)^\s*password\s*=\s*(\S+)\s*$`)
	systemTokenMatch = regexp.MustCompile(`(?m)^\s*system_token\s*=\s*(\S+)\s*$`)
	curlrcMatch      = regexp.MustCompile(`^\s*-*proxy-user[ =]*"(.+):(.+)"\s*$`)
)

// Credentials stores the SCC or service credentials
type Credentials struct {
	Filename    string `json:"file"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	SystemToken string `json:"system_token"`
}

func (c Credentials) String() string {
	return fmt.Sprintf("file: %s, username: %s, password: REDACTED, system_token: %s",
		c.Filename, c.Username, c.SystemToken)
}

func systemCredentialsFile() string {
	return filepath.Join(CFG.FsRoot, globalCredentialsFile)
}

func serviceCredentialsFile(service string) string {
	return filepath.Join(CFG.FsRoot, defaulCredentialsDir, service)
}

func curlrcCredentialsFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// TODO: handle error? log?
		return ""
	}
	return filepath.Join(home, curlrcUserFile)
}

// getCredentials reads the system credentials from the SCCcredentials file
func getCredentials() (Credentials, error) {
	return ReadCredentials(systemCredentialsFile())
}

// ReadCredentials returns the credentials from path
func ReadCredentials(path string) (Credentials, error) {
	Debug.Print("Reading credentials: ", path)
	if !fileExists(path) {
		return Credentials{}, ErrMissingCredentialsFile
	}
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
	tMatch := systemTokenMatch.FindStringSubmatch(string(content))
	if len(uMatch) != 2 || len(pMatch) != 2 {
		return Credentials{}, ErrMalformedSccCredFile
	}
	token := ""
	if len(tMatch) == 2 {
		token = tMatch[1]
	}

	return Credentials{Username: uMatch[1], Password: pMatch[1], SystemToken: token}, nil
}

func (c Credentials) write() error {
	Debug.Print("Writing credentials: ", c)
	path := c.Filename
	if !filepath.IsAbs(path) {
		path = filepath.Join(CFG.FsRoot, defaulCredentialsDir, path)
	}
	dir := filepath.Dir(path)
	if !fileExists(dir) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	buf := bytes.Buffer{}
	fmt.Fprintf(&buf, "username=%s\npassword=%s\n", c.Username, c.Password)
	if c.SystemToken != "" {
		fmt.Fprintf(&buf, "system_token=%s\n", c.SystemToken)
	}
	return os.WriteFile(path, buf.Bytes(), 0600)
}

// CreateCredentials writes credentials to path
func CreateCredentials(login, password, systemToken, path string) error {
	c := Credentials{
		Filename:    path,
		Username:    login,
		Password:    password,
		SystemToken: systemToken,
	}
	return c.write()
}

func writeSystemCredentials(login, password, systemToken string) error {
	path := systemCredentialsFile()
	return CreateCredentials(login, password, systemToken, path)
}

func writeServiceCredentials(serviceName string) error {
	c, err := getCredentials()
	if err != nil {
		return err
	}
	path := serviceCredentialsFile(serviceName)
	// the SystemToken is not written to service credential files
	return CreateCredentials(c.Username, c.Password, "", path)
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

func readCurlrcCredentials(path string) (Credentials, error) {
	Debug.Print("Reading proxy credentials: ", path)
	f, err := os.Open(path)
	if err != nil {
		return Credentials{}, err
	}
	defer f.Close()
	creds, err := parseCurlrcCredentials(f)
	if err != nil {
		return Credentials{}, err
	}
	creds.Filename = path
	Debug.Print("Proxy credentials read: ", creds)
	return creds, nil
}

func parseCurlrcCredentials(r io.Reader) (Credentials, error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		match := curlrcMatch.FindStringSubmatch(line)
		if len(match) == 3 {
			return Credentials{Username: match[1], Password: match[2]}, nil
		}
	}
	return Credentials{}, ErrNoProxyCredentials
}

// ReadCurlrcCredentials reads proxy credentials from default path
func ReadCurlrcCredentials() (Credentials, error) {
	return readCurlrcCredentials(curlrcCredentialsFile())
}
