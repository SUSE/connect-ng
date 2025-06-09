package helpers

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Runner struct {
	t           *testing.T
	commandline []string
	cmd         *exec.Cmd
	stdout      *bytes.Buffer
	stderr      *bytes.Buffer
}

func NewRunner(t *testing.T, commandline string, args ...any) *Runner {
	t.Helper()

	assert := assert.New(t)
	env := NewEnv(t)

	line := fmt.Sprintf(commandline, args...)
	parts := strings.Fields(line)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	if len(parts) < 1 {
		assert.FailNow(fmt.Sprintf("Could not parse commandline `%s`: Check your feature tests", env.RedactRegcodes(line)))
		return nil
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return &Runner{
		t:           t,
		commandline: parts,
		cmd:         cmd,
		stdout:      stdout,
		stderr:      stderr,
	}
}

func (runner *Runner) Run() {
	runner.t.Helper()

	env := NewEnv(runner.t)
	assert := assert.New(runner.t)

	if err := runner.cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			commandline := env.RedactRegcodes(strings.Join(runner.commandline, " "))
			assert.FailNow(fmt.Sprintf("Can not spawn commandline `%s`:", commandline), env.RedactRegcodes(err.Error()))
			return
		}
	}
}

func (runner *Runner) Stdout() string {
	env := NewEnv(runner.t)
	runner.assertCmdHasRun()

	return env.RedactRegcodes(runner.stdout.String())
}

func (runner *Runner) Stderr() string {
	env := NewEnv(runner.t)
	runner.assertCmdHasRun()

	return env.RedactRegcodes(runner.stderr.String())
}

func (runner *Runner) ExitCode() int {
	runner.assertCmdHasRun()
	return runner.cmd.ProcessState.ExitCode()
}

func (runner *Runner) assertCmdHasRun() {
	assert := assert.New(runner.t)
	if runner.cmd.ProcessState.ExitCode() < 0 {
		assert.FailNow("Try to access process results before running the command.")
	}
}
