package connect

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
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
var execute = func(cmd []string, validExitCodes []int) ([]byte, error) {
	Debug.Printf("Executing: %s\n", cmd)
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
	Debug.Printf("Return code: %d\n", exitCode)
	if stdout.Len() > 0 {
		Debug.Println("Output:", stdout.String())
	}
	if stderr.Len() > 0 {
		Debug.Println("Error:", stderr.String())
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

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func removeFile(path string) error {
	Debug.Print("Removing file: ", path)
	if !fileExists(path) {
		return nil
	}
	return os.Remove(path)
}

func isRootFSWritable() bool {
	_, err := execute([]string{"test", "-w", "/"}, []int{zypperOK})
	return err == nil
}

// Cleanup removes system credentials and installed services
func Cleanup() error {
	err := removeSystemCredentials()
	if err != nil {
		return err
	}

	// remove all suse services from zypper
	services, err := installedServices()
	if err != nil {
		return err
	}

	for _, service := range services {
		// NOTE: this check might not work correctly with SMT depending
		//       on the configuration (e.g. listen on https but API
		//       returns URL with http).
		if !strings.Contains(service.URL, CFG.BaseURL) {
			fmt.Printf("%s not in %s\n", CFG.BaseURL, service.URL)
			continue
		}
		if err := removeService(service.Name); err != nil {
			return err
		}

	}
	return nil
}
