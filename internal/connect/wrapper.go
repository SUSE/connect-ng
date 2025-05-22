package connect

import (
	"fmt"

	"github.com/SUSE/connect-ng/internal/collectors"
	"github.com/SUSE/connect-ng/internal/credentials"
	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/SUSE/connect-ng/pkg/registration"
)

// Wrapper is a bridge between API connections via `pkg/connection/` and
// `internal/credentials/`. Use this wrapper in order to perform any API
// requests on the context of SUSEConnect.
type Wrapper struct {
	// Connection is the API connection as defined in `pkg/connection/`.
	Connection *connection.ApiConnection

	// Whether the current system is registered or not. Set after calling `New.`
	Registered bool
}

// Returns a new Wrapper object by taking the given Options into account. Note
// that it will also make an attempt to read any available credentials, and set
// Wrapper.Registered accordingly.
func New(opts *Options) *Wrapper {
	connectionOpts := connection.Options{
		URL:              opts.BaseURL,
		Secure:           !opts.Insecure,
		AppName:          "SUSEConnect",
		Version:          GetShortenedVersion(),
		PreferedLanguage: opts.Language,
		Timeout:          connection.DefaultTimeout,
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
		Connection: connection.New(connectionOpts, creds),
		Registered: registered,
	}
}

// Submit a keepalive request to the server pointed by the configured
// connection.
func (w Wrapper) KeepAlive() error {
	hwinfo, err := FetchSystemInformation()
	if err != nil {
		return fmt.Errorf("could not fetch system's information: %v", err)
	}
	hostname := collectors.FromResult(hwinfo, "hostname", "")

	code, err := registration.Status(w.Connection, hostname, hwinfo, registration.NoExtraData)
	if code != registration.Registered {
		return fmt.Errorf("trying to send a keepalive from a system not yet registered. Register this system first")
	}
	return err
}

func (w Wrapper) Register(regcode string) error {
	hwinfo, err := FetchSystemInformation()
	if err != nil {
		return fmt.Errorf("could not fetch system's information: %v", err)
	}
	hostname := collectors.FromResult(hwinfo, "hostname", "")

	// TODO: do something with the code
	_, err = registration.Register(w.Connection, regcode, hostname, hwinfo, registration.NoExtraData)
	return err
}

// RegisterOrKeepAlive calls either `Register` or `KeepAlive` depending on
// whether the current system is registered or not.
func (w Wrapper) RegisterOrKeepAlive(regcode string) error {
	if w.Registered {
		return w.KeepAlive()
	}
	return w.Register(regcode)
}
