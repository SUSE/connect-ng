package credentials

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/SUSE/connect-ng/internal/util"
)

const (
	DefaultCredentialsDir = "/etc/zypp/credentials.d"
	GlobalCredentialsFile = "/etc/zypp/credentials.d/SCCcredentials"
	CurlrcUserFile        = ".curlrc"
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

// Returns always true, needed to implement the Credentials interface from
// `pkg/`.
func (c Credentials) HasAuthentication() bool {
	return true
}

// Returns the current system token, needed to implement the Credentials
// interface from `pkg/`.
func (c Credentials) Token() (string, error) {
	return c.SystemToken, nil
}

// Writes into the configuration the given system token and updates the value of
// this instance. Needed to implement the Credentials interface from `pkg/`.
func (c Credentials) UpdateToken(token string) error {
	c.SystemToken = token
	return c.write()
}

// Returns the username and password from the system, needed to implement the
// Credentials interface from `pkg/`.
func (c Credentials) Login() (string, string, error) {
	return c.Username, c.Password, nil
}

// Writes into the configuration the given username and password and updates
// their values of this instance. Needed to implement the Credentials interface
// from `pkg/`.
func (c Credentials) SetLogin(username, password string) error {
	c.Username = username
	c.Password = password
	return c.write()
}

func (c Credentials) String() string {
	return fmt.Sprintf("file: %s, username: %s, password: REDACTED, system_token: %s",
		c.Filename, c.Username, c.SystemToken)
}

func SystemCredentialsPath(fsRoot string) string {
	return filepath.Join(fsRoot, GlobalCredentialsFile)
}

func ServiceCredentialsPath(service string, fsRoot string) string {
	return filepath.Join(fsRoot, DefaultCredentialsDir, service)
}

func CurlrcCredentialsPath() string {
	return filepath.Join(util.CurrentHomeDir(), CurlrcUserFile)
}

// ReadCredentials returns the credentials from path
func ReadCredentials(path string) (Credentials, error) {
	util.Debug.Print("Reading credentials: ", path)
	if !util.FileExists(path) {
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
	util.Debug.Print("Credentials read: ", creds)
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
	util.Debug.Print("Writing credentials: ", c)
	path := c.Filename
	dir := filepath.Dir(path)
	if !util.FileExists(dir) {
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

func readCurlrcCredentials(path string) (Credentials, error) {
	util.Debug.Print("Reading proxy credentials: ", path)
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
	util.Debug.Print("Proxy credentials read: ", creds)
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
	return readCurlrcCredentials(CurlrcCredentialsPath())
}
