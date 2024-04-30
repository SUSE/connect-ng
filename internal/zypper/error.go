package zypper

import (
	"errors"
	"fmt"
	"strings"
)

// ZypperError is returned by zypperRun on error
type ZypperError struct {
	Commmand []string
	ExitCode int
	Output   []byte
	Err      error
}

func (ze ZypperError) Error() string {
	return fmt.Sprintf("command '%s' failed\nError: zypper returned %d with '%s' (%s)",
		strings.Join(ze.Commmand, " "), ze.ExitCode, ze.Output, ze.Err)
}

var ErrCannotDetectBaseProduct = errors.New("unable to detect base product")
