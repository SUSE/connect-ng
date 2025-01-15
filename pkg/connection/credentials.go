package connection

// Credentials interface is to be implemented by any struct that knows how to
// save/restore credentials for a given system.
type Credentials interface {
	// Returns true if we can authenticate at all, false otherwise.
	HasAuthentication() bool

	// Token returns the current system used to detect duplicated systems. This
	// token gets rotated on each non read operation.
	Token() (string, error)

	// UpdateToken is called when a token has changed
	UpdateToken(string) error

	// Login returns the username and password
	Login() (string, string, error)

	// SetLogin updates the saved username and password
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
