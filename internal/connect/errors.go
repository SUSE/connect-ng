package connect

import (
	"errors"
	"fmt"

	"github.com/SUSE/connect-ng/pkg/connection"
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

// This method converts connection.ApiError into connect.APIError to not have to deal
// with the same error classes with different namespaces.
func ToAPIError(in error) error {
	if in == nil {
		return nil
	}

	if err, ok := in.(*connection.ApiError); ok {
		return APIError{
			Message: err.Error(),
			Code:    err.Code,
		}
	}

	return APIError{
		Message: in.Error(),
		Code:    -1,
	}
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
