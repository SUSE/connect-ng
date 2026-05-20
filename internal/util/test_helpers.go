package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
)

func ReadTestFile(name string, t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("../../testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestContentMatches(t *testing.T, expected string, got string) {
	if expected != got {
		message := []string{"write: Expected content to match:",
			"---",
			"%s",
			"---",
			"but got:",
			"---",
			"%s",
			"---"}
		t.Errorf(strings.Join(message, "\n"), expected, got)
	}
}

//
// Helpers for mocking Execute()
//

type MockExecutor struct {
	mock.Mock
}

func NewMockExecutor() *MockExecutor {
	return &MockExecutor{}
}

// Setup replaces the standard Execute with a MockExecutor.Execute and
// returns a teardown function that should be called as a deferred call
// to restore the original Execute.
//
// Example:
//
//	mockExecutor := util.NewMockExecutor()
//	someCall := mockExecutor.OnExecuteReturn(
//	    []string{"some", "command", "line"],
//	    []int{0},
//	    []byte("output from running command"),
//	    nil, // no error
//	)
//	... // define additional mock calls
//	teardown := mockExecutor.Setup()
//	defer teardown()
//
//	res, err := CodeThatCallsExecuteWithAboveCommandline()
func (m *MockExecutor) Setup(t *testing.T) func() {
	// retrieve the original Execute
	origExecute := Execute

	// define a teardown function
	teardown := func() {
		// restore the original Execute handler
		Execute = origExecute

		// check that expectations were met, such as calls being made
		// or only
		m.Mock.AssertExpectations(t)
	}

	// replace the original Execute with the Mock handler
	Execute = m.Execute

	return teardown
}

// Execute attempts to retrieve an associated pre-registered mocking
// for the provided command line, and can be used to mock the Execute()
// routine.
func (m *MockExecutor) Execute(cmd []string, validExitCodes []int) (output []byte, err error) {
	// attempt to retrieve any return values previously registered for this
	// call signature; will panic if matching call signature not registered
	args := m.Called(cmd, validExitCodes)

	// extract output return value if provided
	if args.Get(0) != nil {
		output = args.Get(0).([]byte)
	}

	// extract error return value if provided
	if args.Get(1) != nil {
		err = args.Get(1).(error)
	}

	return
}

// compile time validation that MockExecutor.Execute matches ExecuteFunc
var _ ExecuteFunc = (*MockExecutor)(nil).Execute

// add an On("Execute", ...) mocking directive, returning a mock.Call that
// can be manipulated with standard mock.Call methods, such as Once(), Maybe(),
// etc. See https://pkg.go.dev/github.com/stretchr/testify/mock#Call for more
// details on mock.Call.
func (m *MockExecutor) OnExecute(cmd []string, validExitCodes []int) *mock.Call {
	return m.On("Execute", cmd, validExitCodes)
}

// simplified wrapper to quickly setup a simple call with return values
func (m *MockExecutor) OnExecuteReturn(cmd []string, validExitCodes []int, output []byte, err error) *mock.Call {
	return m.OnExecute(cmd, validExitCodes).Return(output, err)
}
