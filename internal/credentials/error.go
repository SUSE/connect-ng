package credentials

import "errors"

// errors
var (
	ErrNoProxyCredentials     = errors.New("unable to read proxy credentials")
	ErrMalformedSccCredFile   = errors.New("cannot parse credentials file")
	ErrMissingCredentialsFile = errors.New("credentials file is missing")
)
