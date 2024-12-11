package connection

// Credentials interface is to be implemented by any struct that knows how to
// save/restore credentials for a given system.
type Credentials interface {
	// Returns true if we can authenticate at all, false otherwise.
	HasAuthentication() bool

	// Returns the authentication triplet in this order: username, password and
	// system token; or an error if the implementing structure cannot return
	// this triplet for whatever reason.
	Triplet() (string, string, string, error)

	// Initializes the credentials structure.
	Load() error

	// Updates the credentials information with the given authentication
	// triplet.
	Update(string, string, string) error
}

// NoCredentials is an empty implementation of the Credentials interface which
// simply returns false to the `HasAuthentication` function. Useful for building
// up connections to API resources which don't require authentication.
type NoCredentials struct{}

func (NoCredentials) HasAuthentication() bool {
	return false
}

func (NoCredentials) Triplet() (string, string, string, error) {
	return "", "", "", nil
}

func (NoCredentials) Load() error {
	return nil
}

func (NoCredentials) Update(string, string, string) error {
	return nil
}
