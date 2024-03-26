package credentials

import "errors"

// errors
var (
	ErrNoProxyCredentials     = errors.New("Unable to read proxy credentials")
	ErrMalformedSccCredFile   = errors.New("Cannot parse credentials file")
	ErrMissingCredentialsFile = errors.New("Credentials file is missing")
)
