package connect

import (
	"errors"
	"fmt"
)

// export errors that package main needs
var (
	ErrSystemNotRegistered        = errors.New("System not registered")
	ErrPingFromUnregistered       = errors.New("Keepalive ping not allowed from unregistered system.")
	ErrBaseProductDeactivation    = errors.New("Unable to deactivate base product")
	ErrListExtensionsUnregistered = errors.New("System not registered")
)

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

func (je JSONError) Unwrap() error {
	return je.Err
}
