package util

import "fmt"

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
