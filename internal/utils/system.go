package utils

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"github.com/SUSE/connect-ng/internal/logging"
)

var systemEcho bool

// SetSystemEcho toggles piping of executed command's outputs to stdout/stderr
// returns true if it was enabled before, false otherwise
func SetSystemEcho(v bool) bool {
	prev := systemEcho
	systemEcho = v
	return prev
}

// Assign function for running external commands to a variable so it can be mocked by tests.
var Execute = func(cmd []string, validExitCodes []int) ([]byte, error) {
	logging.Debug.Print("Executing: ", cmd)
	var stderr, stdout bytes.Buffer
	comm := exec.Command(cmd[0], cmd[1:]...)
	if systemEcho {
		comm.Stdout = io.MultiWriter(os.Stdout, &stdout)
		comm.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		comm.Stdout = &stdout
		comm.Stderr = &stderr
	}
	comm.Env = append(os.Environ(), "LC_ALL=C")
	err := comm.Run()
	exitCode := comm.ProcessState.ExitCode()
	logging.Debug.Printf("Return code: %d\n", exitCode)
	if stdout.Len() > 0 {
		logging.Debug.Print("Output: ", stdout.String())
	}
	if stderr.Len() > 0 {
		logging.Debug.Print("Error: ", stderr.String())
	}
	// TODO Ruby version also checks stderr for "ABORT request"
	if err != nil && !containsInt(validExitCodes, exitCode) {
		output := stderr.Bytes()
		// zypper with formatter option writes to stdout instead of stderr
		if len(output) == 0 {
			output = stdout.Bytes()
		}
		output = bytes.TrimSuffix(output, []byte("\n"))
		ee := ExecuteError{Commmand: cmd, ExitCode: exitCode, Output: output, Err: err}
		return nil, ee
	}
	out := stdout.Bytes()
	out = bytes.TrimSuffix(out, []byte("\n"))
	return out, nil
}

func containsInt(s []int, i int) bool {
	for _, e := range s {
		if e == i {
			return true
		}
	}
	return false
}
