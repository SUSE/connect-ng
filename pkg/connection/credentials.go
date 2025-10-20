package connection

// Credentials interface is to be implemented by any struct that knows how to
// save/restore credentials for a given system.
//
// In SCC each system is identified by a triple of: login, password, and system
// token. The login and password are the standard credentials as expected on any
// service, and the system token is a mechanism that we have introduced in order
// to detect system duplicates.
//
// Every time a client sends a non-read request to the SCC server,
// the server returns an updated system token. Clients must store this
// token and include it in subsequent requests.
//
// In this library, client–server communication is handled at the transport
// layer. This interface defines two hooks that the transport layer uses
// to manage the system token:
//
//   - Token()        // returns the latest stored system token
//   - UpdateToken()  // updates/sets the stored system token with the latest value
type Credentials interface {
	// Returns true if we can authenticate at all, false otherwise.
	HasAuthentication() bool

	// Token() returns the last system token recorded by this client.
	//
	// The system token is rotated after each non-read operation and updated
	// through UpdateToken(). If no token has been recorded yet, the method
	// may return an empty string, indicating that the system has not been
	// registered.
	Token() (string, error)

	// UpdateToken() is a hook used by the transport layer to persist a new
	// system token received after a non-read request.
	//
	// This allows the client to store the updated token so that it can be
	// provided in subsequent requests through the Token() method.
	UpdateToken(string) error

	// Login() returns the username and password associated with the system.
	Login() (string, string, error)

	// SetLogin is a hook used by the transport layer to persist the system’s
	// login and password credentials.
	//
	// The system token is managed separately through UpdateToken().
	SetLogin(string, string) error
}

// NoCredentials is an empty implementation of the Credentials interface which
// simply returns false to the `HasAuthentication` function. Useful for building
// up connections to API resources which don't require authentication.
type NoCredentials struct{}

func (NoCredentials) HasAuthentication() bool {
	return false
}

func (NoCredentials) Token() (string, error) {
	return "", nil
}

func (NoCredentials) UpdateToken(string) error {
	return nil
}

func (NoCredentials) Login() (string, string, error) {
	return "", "", nil
}

func (NoCredentials) SetLogin(string, string) error {
	return nil
}
