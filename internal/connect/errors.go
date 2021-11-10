package connect

import (
	"errors"
	"fmt"
	"strings"
)

// export errors that package main needs
var (
	ErrMalformedSccCredFile       = errors.New("Cannot parse credentials file")
	ErrMissingCredentialsFile     = errors.New("Credentials file is missing")
	ErrSystemNotRegistered        = errors.New("System not registered")
	ErrBaseProductDeactivation    = errors.New("Unable to deactivate base product")
	ErrCannotDetectBaseProduct    = errors.New("Unable to detect base product")
	ErrListExtensionsUnregistered = errors.New("System not registered")
	ErrNoProxyCredentials         = errors.New("Unable to read proxy credentials")
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
	Commmand []string
	ExitCode int
	Output   []byte
}

func (ze ZypperError) Error() string {
	return fmt.Sprintf("command '%s' failed\nError: zypper returned %d with '%s'",
		strings.Join(ze.Commmand, " "), ze.ExitCode, ze.Output)
}

// APIError is returned on failed HTTP requests
type APIError struct {
	Code    int
	Message string
}

func (ae APIError) Error() string {
	return fmt.Sprintf("Error: Registration server returned '%s' (%d)", ae.Message, ae.Code)
}

// JSONError is returned on failed JSON decoding
type JSONError struct {
	Err error
}

func (je JSONError) Error() string {
	return fmt.Sprintf("JSON Error: %v", je.Err)
}
