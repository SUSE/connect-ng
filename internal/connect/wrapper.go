package connect

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/SUSE/connect-ng/internal/collectors"
	"github.com/SUSE/connect-ng/internal/credentials"
	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/SUSE/connect-ng/pkg/labels"
	"github.com/SUSE/connect-ng/pkg/registration"
)

// Wrap everything into an interface so we can mock this calls later on
// when unit testing
type WrappedAPI interface {
	KeepAlive(uptimeTracking bool) error
	Register(regcode, instanceDataFile string) error
	RegisterOrKeepAlive(regcode, instanceDataFile string, uptimeTracking bool) error
	IsRegistered() bool
	AssignLabels(labels []string) ([]labels.Label, error)

	// Return the underlying API object in case an unknown API needs to be
	// implemented in SUSEConnect.
	GetConnection() connection.Connection
}

// Wrapper is a bridge between API connections via `pkg/connection/` and
// `internal/credentials/`. Use this wrapper in order to perform any API
// requests on the context of SUSEConnect.
type Wrapper struct {
	// Connection is the API connection as defined in `pkg/connection/`.
	Connection connection.Connection

	// Whether the current system is registered or not. Set after calling `New.`
	Registered bool
}

// Returns true if proxy setup is enabled at the system level. This is specific
// to SUSE.
func proxyEnabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("PROXY_ENABLED")))

	// NOTE: if the value is not set, we return true so Go figures this out.
	return value == "" || value == "y" || value == "yes" || value == "t" || value == "true"
}

// Returns the proxy setup if needed.
func proxyWithAuth(req *http.Request) (*url.URL, error) {
	// Check for the special "PROXY_ENABLED" environment variable which might be
	// set in a SUSE system. If it is set to a falsey value, then we skip proxy
	// detection regardless of other environment variables.
	if !proxyEnabled() {
		return nil, nil
	}

	// A nil proxyURL implies that the proxy setup has been explicitly disabled.
	proxyURL, err := http.ProxyFromEnvironment(req)
	if proxyURL == nil || err != nil {
		return proxyURL, err
	}
	// Add or replace proxy credentials if configured
	if c, err := credentials.ReadCurlrcCredentials(); err == nil {
		proxyURL.User = url.UserPassword(c.Username, c.Password)
	}
	return proxyURL, nil
}

// Returns a new Wrapper object by taking the given Options into account. Note
// that it will also make an attempt to read any available credentials, and set
// Wrapper.Registered accordingly.
func NewWrappedAPI(opts *Options) WrappedAPI {
	connectionOpts := connection.Options{
		URL:              opts.BaseURL,
		Secure:           !opts.Insecure,
		AppName:          "SUSEConnect",
		Version:          GetShortenedVersion(),
		PreferedLanguage: opts.Language,
		Timeout:          connection.DefaultTimeout,
		Proxy:            proxyWithAuth,
	}

	credentialsPath := credentials.SystemCredentialsPath(opts.FsRoot)
	creds, err := credentials.ReadCredentials(credentialsPath)
	registered := false
	if err == nil {
		registered = true
	} else {
		// If the credentials could not be read, then probably it was because
		// they did not exist. In this case, initialize at least the filename so
		// future writes don't fail and can create a new credentials file.
		creds.Filename = credentialsPath
	}

	return &Wrapper{
		Connection: connection.New(connectionOpts, &creds),
		Registered: registered,
	}
}

// Submit a keepalive request to the server pointed by the configured
// connection.
func (w Wrapper) KeepAlive(uptimeTracking bool) error {
	hwinfo, err := FetchSystemInformation()
	if err != nil {
		return fmt.Errorf("could not fetch system's information: %v", err)
	}
	hostname := collectors.FromResult(hwinfo, "hostname", "")

	// If the uptime tracking log is requested via the configuration, attach it
	// now.
	extraData := registration.NoExtraData
	if uptimeTracking {
		data, err := readUptimeLogFile(UptimeLogFilePath)
		if err != nil {
			return err
		}
		extraData["online_at"] = data
	}

	code, err := registration.Status(w.Connection, hostname, hwinfo, extraData)
	if code != registration.Registered {
		return fmt.Errorf("trying to send a keepalive from a system not yet registered. Register this system first")
	}
	return err
}

func (w Wrapper) Register(regcode, instanceDataFile string) error {
	hwinfo, err := FetchSystemInformation()
	if err != nil {
		return fmt.Errorf("could not fetch system's information: %v", err)
	}
	hostname := collectors.FromResult(hwinfo, "hostname", "")

	// If an instance-data file was provided, try to read it and attach it as
	// "extra" data. This will be used inside of the `registration.Register`
	// code.
	extraData := registration.NoExtraData
	if instanceDataFile != "" {
		data, err := os.ReadFile(instanceDataFile)
		if err != nil {
			return err
		}
		extraData["instance_data"] = string(data)
	}

	// NOTE: we are not interested in the code. Hence, we don't save it
	// anywhere.
	_, err = registration.Register(w.Connection, regcode, hostname, hwinfo, extraData)
	return err
}

// RegisterOrKeepAlive calls either `Register` or `KeepAlive` depending on
// whether the current system is registered or not.
func (w Wrapper) RegisterOrKeepAlive(regcode, instanceDataFile string, uptimeTracking bool) error {
	if w.Registered {
		return w.KeepAlive(uptimeTracking)
	}
	return w.Register(regcode, instanceDataFile)
}

func (w Wrapper) IsRegistered() bool {
	return w.Registered
}

func (w Wrapper) GetConnection() connection.Connection {
	return w.Connection
}

func (w Wrapper) AssignLabels(assigned []string) ([]labels.Label, error) {
	collection := []labels.Label{}

	for _, name := range assigned {
		name = strings.TrimSpace(name)
		collection = append(collection, labels.Label{Name: name})
	}
	util.Debug.Printf(util.Bold("Setting Labels %s"), strings.Join(assigned, ","))

	return labels.AssignLabels(w.Connection, collection)
}
