package connect

import (
	"errors"
	"fmt"
)

var (
	ErrMalformedSccCredFile       = errors.New("Unable to parse credentials")
	ErrMissingCredentialsFile     = errors.New("Credentials file is missing")
	ErrSystemNotRegistered        = errors.New("System not registered")
	ErrBaseProductDeactivation    = errors.New("Unable to deactivate base product")
	ErrCannotDetectBaseProduct    = errors.New("Unable to detect base product")
	ErrListExtensionsUnregistered = errors.New("System not registered")
)

// ExecuteError is returned from execute() on error
type ExecuteError struct {
	Commmand []string
	ExitCode int
	Output   []byte
	Err      error
}

func (ee ExecuteError) Error() string {
	return fmt.Sprintf(
		"ExecuteError: Cmd: %s, RC: %d, Error: %s, Output: %s",
		ee.Commmand, ee.ExitCode, ee.Err, ee.Output)
}

// ZypperError is returned by zypperRun on error
type ZypperError struct {
	ExitCode int
	Output   []byte
}

func (ze ZypperError) Error() string {
	return fmt.Sprintf("Error: zypper returned %d with '%s'", ze.ExitCode, ze.Output)
}

// APIError is returned on failed HTTP requests
type APIError struct {
	Code    int
	Message string
}

func (ae APIError) Error() string {
	return fmt.Sprintf("Error: Registration server returned '%s' (%d)", ae.Message, ae.Code)
}
