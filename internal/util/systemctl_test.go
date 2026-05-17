package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type systemctlTestCfg struct {
	SubCmd    string
	Args      []string
	Output    []byte
	Error     error
	ExitCodes []int
}

func systemctlSimpleTestMockHandler(cmd []string, exitCodes []int, env *MockExecuteEnv) ([]byte, error) {
	cfg := env.Cfg.(*systemctlTestCfg)
	return cfg.Output, cfg.Error
}

func TestSystemctlNormalOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test setup
	testCfg := &systemctlTestCfg{
		SubCmd:    "testSubCmd",
		Args:      []string{"testArg1", "testArg2"},
		Output:    []byte("test output"),
		Error:     nil,
		ExitCodes: []int{0},
	}

	mockEnv := NewExecuteMockEnv(testCfg)
	mockEnv.Setup(systemctlSimpleTestMockHandler)
	defer mockEnv.Teardown()

	// construct expected output and call command line and exit codes
	expectedCmdLine := append([]string{}, SystemctlBaseCmd...)
	expectedCmdLine = append(expectedCmdLine, testCfg.SubCmd)
	expectedCmdLine = append(expectedCmdLine, testCfg.Args...)
	expectedExitCodes := []int{0}
	expectedOutput := []byte(testCfg.Output)

	// run test case
	output, err := Systemctl(testCfg.SubCmd, testCfg.Args...)

	// check returned values
	assert.Equal(expectedOutput, output, "Systemctl() returned unexpected output")
	assert.NoError(err, "Systemctl() returned unexpected error")

	// check that Execute() was called as expected
	require.Equal(1, mockEnv.NumCalls, "Systemctl() called Execute() an unexpected number of times")
	assert.Equal(expectedCmdLine, mockEnv.CmdLinesList[0], "Systemctl() called Execute() with unexpected cmd argument")
	assert.Equal(expectedExitCodes, mockEnv.ExitCodesList[0], "Systemctl() called Execute() with unexpected exitCodes argument")
}

func TestSystemctlFailedOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test setup
	testCfg := &systemctlTestCfg{
		SubCmd:    "testSubCmd",
		Args:      []string{"testArg1", "testArg2"},
		Output:    []byte(nil),
		Error:     fmt.Errorf("test error"),
		ExitCodes: []int{0},
	}

	// setup mockingt for Execute
	mockEnv := NewExecuteMockEnv(testCfg)
	mockEnv.Setup(systemctlSimpleTestMockHandler)
	defer mockEnv.Teardown()

	// construct expected output and call command line and exit codes
	expectedCmdLine := append([]string{}, SystemctlBaseCmd...)
	expectedCmdLine = append(expectedCmdLine, testCfg.SubCmd)
	expectedCmdLine = append(expectedCmdLine, testCfg.Args...)
	expectedExitCodes := []int{0}

	// run test case
	output, err := Systemctl(testCfg.SubCmd, testCfg.Args...)

	// check returned values
	assert.Nil(output, "Systemctl() should have returned nil for output")
	assert.Error(err, "Systemctl() should have returned an error")
	assert.ErrorIs(err, testCfg.Error, "Systemctl() returned unexpected error")

	// check that Execute() was called as expected
	require.Equal(1, mockEnv.NumCalls, "Systemctl() called Execute() an unexpected number of times")
	assert.Equal(expectedCmdLine, mockEnv.CmdLinesList[0], "Systemctl() called Execute() with unexpected cmd argument")
	assert.Equal(expectedExitCodes, mockEnv.ExitCodesList[0], "Systemctl() called Execute() with unexpected exitCodes argument")
}
