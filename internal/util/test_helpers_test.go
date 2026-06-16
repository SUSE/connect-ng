package util

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test the MockExecute helper functionality

func TestMockExecutor(t *testing.T) {
	// remember the original Execute
	originalExecute := Execute

	// define a test config that will be used by test handler
	type testCfg struct {
		name      string
		cmd       []string
		exitCodes []int
		output    []byte
		err       error
	}

	// define a test error
	testError := assert.AnError

	// setup a table of subtests to run
	for _, tc := range []testCfg{
		{
			name:      "verify successful command handling",
			cmd:       []string{"success"},
			exitCodes: []int{0},
			output:    []byte("successful command execution"),
			err:       nil,
		},
		{
			name:      "verify unpermitted failed command handling",
			cmd:       []string{"failure"},
			exitCodes: []int{0},
			output:    []byte(nil),
			err:       testError,
		},
		{
			name:      "verify permitted failing command handling",
			cmd:       []string{"failure"},
			exitCodes: []int{0, 1},
			output:    []byte("permitted failing command execution"),
			err:       nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// general setup
			assert := assert.New(t)
			require := require.New(t)

			// create the MockExec and verify that cfg was setup correctly
			mockExec := NewMockExecutor()

			// setup the mock Execute() which returns a teardown function for use with defer
			teardown := mockExec.Setup(t)

			// verify that Execute was replaced by a mock Execute
			require.NotNil(Execute, "MockExecutor.Setup() did not mock Execute correctly")
			require.NotEqual(
				reflect.ValueOf(originalExecute).Pointer(),
				reflect.ValueOf(Execute).Pointer(),
				"MockExecutor.Setup() did not replace the original Execute function properly",
			)

			require.Equal(
				reflect.ValueOf(mockExec.Execute).Pointer(),
				reflect.ValueOf(Execute).Pointer(),
				"MockExecutor.Setup() did setup the mock Execute function properly",
			)

			// setup mocking for the requested call to happen once only
			mockExec.OnExecute(tc.cmd, tc.exitCodes).Return(tc.output, tc.err).Once()

			// perform the mocked call
			output, err := Execute(tc.cmd, tc.exitCodes)

			// verify that expected mocked output and error were returned
			assert.Equal(tc.output, output, "Mocked Execute() returned unexpected output")
			assert.Equal(tc.err, err, "Mocked Execute() returned unexpected error")

			// verify that after teardown the original Execute is restored
			teardown()
			require.NotNil(Execute, "MockExecutor.Setup()'s teardown function did not retore Execute correctly")
			require.Equal(
				reflect.ValueOf(originalExecute).Pointer(),
				reflect.ValueOf(Execute).Pointer(),
				"MockExecutor.Setup()'s teardown function did not restore the original Execute function",
			)

		})
	}
}
