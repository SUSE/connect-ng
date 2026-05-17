package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

type MockExecuteEnv struct {
	// Public Attributes
	NumCalls      int
	CmdLinesList  [][]string
	ExitCodesList [][]int
	Cfg           any

	// Private Attributes
	origExecute ExecuteFunc
}

func NewExecuteMockEnv(cfg any) *MockExecuteEnv {
	return &MockExecuteEnv{
		NumCalls:      0,
		CmdLinesList:  [][]string{},
		ExitCodesList: [][]int{},
		Cfg:           cfg,
		origExecute:   nil,
	}
}

type MockExecuteFunc func(cmd []string, exitCodes []int, env *MockExecuteEnv) ([]byte, error)

func (m *MockExecuteEnv) Setup(handler MockExecuteFunc) {
	// save the original Execute
	m.origExecute = Execute
	Execute = func(cmd []string, exitCodes []int) ([]byte, error) {
		// the mocked execute was called so update tracking settings
		m.NumCalls++
		m.CmdLinesList = append(m.CmdLinesList, cmd)
		m.ExitCodesList = append(m.ExitCodesList, exitCodes)

		// call the handler, passing in the cmd, exitCodes and mock env
		// and return the results
		return handler(cmd, exitCodes, m)
	}
}

func (m *MockExecuteEnv) OriginalExecute() ExecuteFunc {
	return m.origExecute
}

func (m *MockExecuteEnv) Teardown() {
	// restore the original Execute
	Execute = m.origExecute
	m.origExecute = nil
}
