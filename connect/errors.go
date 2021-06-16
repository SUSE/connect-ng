package connect

import (
	"errors"
)

var (
	ErrMalformedSccCredFile = errors.New("Unable to parse credentials")
	ErrSystemNotRegistered  = errors.New("System not registered")
)
