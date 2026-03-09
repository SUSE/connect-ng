package util

import (
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test the MockExecute helper functionality

func TestMockExecuteEnv(t *testing.T) {
	// remember the original Execute
	originalExecute := Execute

	// define a test config that will be used by test handler
	type testCfg struct {
		name      string
		cmd       []string
		exitCodes []int
		status    int
		output    []byte
		err       error
		numCalls  int
	}

	// define a test error
	testError := fmt.Errorf("test error")

	// define a handler that uses the defined config and returns specific values
	var handler MockExecuteFunc = func(cmd []string, exitCodes []int, env *MockExecuteEnv) ([]byte, error) {
		cfg := env.Cfg.(*testCfg)

		var output []byte
		var err error

		switch cmd[0] {
		case "success":
			output = []byte("successful command execution")
		case "failure":
			if slices.Contains(exitCodes, cfg.status) {
				output = []byte("permitted failing command execution")
			} else {
				err = testError
			}
		}

		return output, err
	}

	// setup a table of subtests to run
	for _, tc := range []testCfg{
		{
			name:      "verify successful command handling",
			cmd:       []string{"success"},
			exitCodes: []int{0},
			status:    0,
			output:    []byte("successful command execution"),
			err:       nil,
			numCalls:  1,
		},
		{
			name:      "verify unpermitted failed command handling",
			cmd:       []string{"failure"},
			exitCodes: []int{0},
			status:    1,
			output:    []byte(nil),
			err:       testError,
			numCalls:  1,
		},
		{
			name:      "verify permitted failing command handling",
			cmd:       []string{"failure"},
			exitCodes: []int{0, 1},
			status:    1,
			output:    []byte("permitted failing command execution"),
			err:       nil,
			numCalls:  1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// general setup
			assert := assert.New(t)
			require := require.New(t)

			// use the provided test config
			cfg := &tc

			// create the MockEnv and verify that cfg was setup correctly
			mockEnv := NewExecuteMockEnv(cfg)
			assert.Equal(cfg, mockEnv.Cfg, "MockExecuteEnv.Cfg not setup correctly")

			// setup mocking of Execute
			mockEnv.Setup(handler)

			// verify that Execute was replaced by a mock Execute
			require.NotNil(Execute, "MockExecuteEnv.Setup() did not correctly mock the Execute routine")
			if reflect.ValueOf(Execute).Pointer() == reflect.ValueOf(originalExecute).Pointer() {
				t.Fatal("MockExecuteEnv.Setup() did not replace the original Execute function properly")
			}

			// verify that the original Execute was saved when mocking was setup
			require.NotNil(mockEnv.OriginalExecute(), "MockExecuteEnv.Setup() did not correctly save the original Execute routine")
			if reflect.ValueOf(mockEnv.OriginalExecute()).Pointer() != reflect.ValueOf(originalExecute).Pointer() {
				t.Fatal("MockExecuteEnv.Setup() did not save the original Execute function properly")
			}

			// perform the mocked call
			output, err := Execute(cfg.cmd, cfg.exitCodes)

			// verify that after teardown the original Execute is restored
			mockEnv.Teardown()
			require.NotNil(Execute, "MockExecuteEnv.Teardown() did not correctly restore the Execute routine")
			if reflect.ValueOf(Execute).Pointer() != reflect.ValueOf(originalExecute).Pointer() {
				t.Fatal("MockExecuteEnv.Teardown() did not restore the original Execute function properly")
			}

			// verify that expected mocked output and error were returned
			assert.Equal(cfg.output, output, "Mocked Execute() returned unexpected output")
			assert.Equal(cfg.err, err, "Mocked Execute() returned unexpected error")

			// verify that Execute() was called as expected
			require.Equal(cfg.numCalls, mockEnv.NumCalls, "Execute() called unexpected number of times")
			assert.Equal(cfg.cmd, mockEnv.CmdLinesList[0], "Execute() called with unexpected cmd argument")
			assert.Equal(cfg.exitCodes, mockEnv.ExitCodesList[0], "Execute() called with unexpected exitCodes argument")

		})
	}
}
