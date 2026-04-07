package util

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListUnitFilesOutputMatcher(t *testing.T) {
	testUnitFile := "test.unit"
	enabled := "enabled"

	testCases := []struct {
		name          string
		matcher       *regexp.Regexp
		input         string
		fieldMatches  [][]byte
		expectedCount int
	}{
		//
		// systemctl list-unit-files output matcher
		//
		{
			name:          "list-unit-files output matcher for valid output (unitfile state preset)",
			matcher:       listUnitFilesOutputMatcher,
			input:         fmt.Sprintf("%s %s %s", testUnitFile, enabled, enabled),
			fieldMatches:  [][]byte{[]byte(testUnitFile), []byte(enabled)},
			expectedCount: 3, // line + 2 fields
		},
		{
			name:          "list-unit-files output matcher for valid output on older systemd versions like SLE-12SP5 (unitfile state)",
			matcher:       listUnitFilesOutputMatcher,
			input:         fmt.Sprintf("%s %s", testUnitFile, enabled),
			fieldMatches:  [][]byte{[]byte(testUnitFile), []byte(enabled)},
			expectedCount: 3, // line + 2 fields
		},
		{
			name:          "list-unit-files output matcher for invalid output (missing unitfile)",
			matcher:       listUnitFilesOutputMatcher,
			input:         fmt.Sprintf(" %s", enabled),
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},
		{
			name:          "list-unit-files output matcher for invalid output (missing state)",
			matcher:       listUnitFilesOutputMatcher,
			input:         testUnitFile,
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},
		{
			name:          "list-unit-files output matcher for invalid output (empty)",
			matcher:       listUnitFilesOutputMatcher,
			input:         "",
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			matches := tc.matcher.FindSubmatch([]byte(tc.input))
			assert.Len(matches, tc.expectedCount, "FindSubMatch(%s) for %q didn't return expected number of matches", tc.matcher, tc.input)
			if len(matches) == 0 { // skip matched field checks when none expected
				return
			}
			assert.Equal([]byte(tc.input), matches[0], "FindSubmatch(%s) for %q should have matched entire input string", tc.matcher, tc.input)
			assert.Equal(tc.fieldMatches, matches[1:], "FindSubmatch(%s) for %q should have matched expected fields", tc.matcher, tc.input)
		})
	}
}

func TestListMatchingUnitFilesOfTypeAndStateNormalOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test setup
	testSvc := "test.service"
	testSvcMatcher := regexp.MustCompile(fmt.Sprintf("^%s$", testSvc))
	testPattern := "test*"
	testCfg := &systemctlTestCfg{
		SubCmd: "list-unit-files",
		Args:   []string{"--type", "service", "--state", "enabled", testPattern},
		Output: []byte(strings.Join(
			[]string{
				fmt.Sprintf("%s enabled enabled", testSvc),
			},
			"\n",
		)),
		Error:     nil,
		ExitCodes: []int{0},
	}

	mockEnv := NewExecuteMockEnv(testCfg)
	mockEnv.Setup(systemctlSimpleTestMockHandler)
	defer mockEnv.Teardown()

	// construct expected output, systemctl call command line and exit codes
	expectedCmdLine := append([]string{}, SystemctlBaseCmd...)
	expectedCmdLine = append(expectedCmdLine, testCfg.SubCmd)
	expectedCmdLine = append(expectedCmdLine, testCfg.Args...)
	expectedExitCodes := []int{0}
	expectedUnitFiles := []*SystemdUnitFile{
		NewSystemdUnitFile(testSvc),
	}

	// run test case
	unitFiles, err := ListMatchingUnitFilesOfTypeAndState("test*", "service", "enabled", testSvcMatcher)

	// check returned values
	assert.NotNil(unitFiles, "ListMatchingUnitFilesOfTypeAndState() returned nil for unitFiles list")
	assert.NoError(err, "ListMatchingUnitFilesOfTypeAndState() returned unexpected error")
	assert.Equal(expectedUnitFiles, unitFiles, "ListMatchingUnitFilesOfTypeAndState() returned unexpected output")

	// check Execute() was called as expected
	require.Equal(1, mockEnv.NumCalls, "Execute() called unexpected number of times")
	assert.Equal(expectedCmdLine, mockEnv.CmdLinesList[0], "listMatchingUnitFilesOfTypeAndState() called util.Execute() with unexperect cmd argument")
	assert.Equal(expectedExitCodes, mockEnv.ExitCodesList[0], "Systemctl() called Execute() with unexpected exitCodes argument")
}

func TestListMatchingUnitFilesOfTypeAndStateNoMatchesOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test setup
	testSvc := "test.service"
	testSvcMatcher := regexp.MustCompile(fmt.Sprintf("^%s$", testSvc))
	testPattern := "test*"
	testCfg := &systemctlTestCfg{
		SubCmd:    "list-unit-files",
		Args:      []string{"--type", "service", "--state", "enabled", testPattern},
		Output:    []byte(strings.Join([]string{}, "\n")),
		Error:     nil,
		ExitCodes: []int{0},
	}

	mockEnv := NewExecuteMockEnv(testCfg)
	mockEnv.Setup(systemctlSimpleTestMockHandler)
	defer mockEnv.Teardown()

	// construct expected output, systemctl call command line and exit codes
	expectedCmdLine := append([]string{}, SystemctlBaseCmd...)
	expectedCmdLine = append(expectedCmdLine, testCfg.SubCmd)
	expectedCmdLine = append(expectedCmdLine, testCfg.Args...)
	expectedExitCodes := []int{0}

	// run test case
	unitFiles, err := ListMatchingUnitFilesOfTypeAndState("test*", "service", "enabled", testSvcMatcher)

	// check returned values
	assert.Empty(unitFiles, "ListMatchingUnitFilesOfTypeAndState() should have returned an empty list of unit files")
	assert.NoError(err, "ListMatchingUnitFilesOfTypeAndState() returned unexpected error")

	// check Execute() was called as expected
	require.Equal(1, mockEnv.NumCalls, "Execute() called unexpected number of times")
	assert.Equal(expectedCmdLine, mockEnv.CmdLinesList[0], "listMatchingUnitFilesOfTypeAndState() called util.Execute() with unexperect cmd argument")
	assert.Equal(expectedExitCodes, mockEnv.ExitCodesList[0], "Systemctl() called Execute() with unexpected exitCodes argument")
}

func TestListMatchingUnitFilesOfTypeAndStateFailingSystemCtl(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test setup
	testSvc := "test.service"
	testSvcMatcher := regexp.MustCompile(fmt.Sprintf("^%s$", testSvc))
	testPattern := "test*"
	testError := fmt.Errorf("test error")
	testCfg := &systemctlTestCfg{
		SubCmd:    "list-unit-files",
		Args:      []string{"--type", "service", "--state", "enabled", testPattern},
		Output:    []byte(nil),
		Error:     testError,
		ExitCodes: []int{0},
	}

	mockEnv := NewExecuteMockEnv(testCfg)
	mockEnv.Setup(systemctlSimpleTestMockHandler)
	defer mockEnv.Teardown()

	// construct expected output, systemctl call command line and exit codes
	expectedCmdLine := append([]string{}, SystemctlBaseCmd...)
	expectedCmdLine = append(expectedCmdLine, testCfg.SubCmd)
	expectedCmdLine = append(expectedCmdLine, testCfg.Args...)
	expectedExitCodes := []int{0}

	// run test case
	unitFiles, err := ListMatchingUnitFilesOfTypeAndState("test*", "service", "enabled", testSvcMatcher)

	// check returned values
	assert.Nil(unitFiles, "ListMatchingUnitFilesOfTypeAndState() should have returned a null list of unit files")
	assert.Error(err, "ListMatchingUnitFilesOfTypeAndState() should have returned an error")
	assert.ErrorIs(err, testError, "ListMatchingUnitFilesOfTypeAndState() should have returned an instance of testError")

	// check Execute() was called as expected
	require.Equal(1, mockEnv.NumCalls, "Execute() called unexpected number of times")
	assert.Equal(expectedCmdLine, mockEnv.CmdLinesList[0], "listMatchingUnitFilesOfTypeAndState() called util.Execute() with unexperect cmd argument")
	assert.Equal(expectedExitCodes, mockEnv.ExitCodesList[0], "Systemctl() called Execute() with unexpected exitCodes argument")
}

func TestUnitFilePropertyNormalOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	testSvc := "test.service"
	testCfg := &systemctlTestCfg{
		SubCmd:    "show",
		Args:      []string{"--property", "ExecStart", "test.service"},
		Output:    []byte("test output"),
		Error:     nil,
		ExitCodes: []int{0},
	}
	mockEnv := NewExecuteMockEnv(testCfg)

	// expected test values
	expectedCmdLine := append([]string{}, SystemctlBaseCmd...)
	expectedCmdLine = append(expectedCmdLine, testCfg.SubCmd)
	expectedCmdLine = append(expectedCmdLine, testCfg.Args...)
	expectedNumCalls := 1

	// setup util.Execute() mocking
	mockEnv.Setup(systemctlSimpleTestMockHandler)
	defer mockEnv.Teardown()

	// run test case
	uf := NewSystemdUnitFile(testSvc)
	output, err := uf.Property("ExecStart")

	// check returned values
	assert.Equal(testCfg.Output, output, "SystemdUnitFile.Property() returned unexpected output")
	assert.NoError(err, "SystemdUnitFile.Property() returned unexpected error")

	// check Execute() was called as expected
	require.Equal(expectedNumCalls, mockEnv.NumCalls, "Execute() called unexpected number of times")
	assert.Equal(expectedCmdLine, mockEnv.CmdLinesList[0], "SystemdUnitFile.Property() triggered Execute() called with unexpected cmd argument")
	assert.Equal(testCfg.ExitCodes, mockEnv.ExitCodesList[0], "SystemdUnitFile.Property() triggered util.Execute() called with unexpected exit codes")

}

func TestUnitFilePropertyFailedOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	testSvc := "test.service"
	testError := fmt.Errorf("test error")
	testCfg := &systemctlTestCfg{
		SubCmd:    "show",
		Args:      []string{"--property", "ExecStart", "test.service"},
		Output:    []byte(nil),
		Error:     testError,
		ExitCodes: []int{0},
	}
	mockEnv := NewExecuteMockEnv(testCfg)

	// expected test values
	expectedCmdLine := append([]string{}, SystemctlBaseCmd...)
	expectedCmdLine = append(expectedCmdLine, testCfg.SubCmd)
	expectedCmdLine = append(expectedCmdLine, testCfg.Args...)
	expectedNumCalls := 1

	// setup util.Execute() mocking
	mockEnv.Setup(systemctlSimpleTestMockHandler)
	defer mockEnv.Teardown()

	// run test case
	uf := NewSystemdUnitFile(testSvc)
	output, err := uf.Property("ExecStart")

	// check returned values
	assert.Equal(testCfg.Output, output, "SystemdUnitFile.Property() returned unexpected output")
	assert.Error(err, "SystemdUnitFile.Property() should have returned an error")
	assert.ErrorIs(err, testError, "SystemdUnitFile.Property() should have returned an instance of testError")

	// check Execute() was called as expected
	require.Equal(expectedNumCalls, mockEnv.NumCalls, "Execute() called unexpected number of times")
	assert.Equal(expectedCmdLine, mockEnv.CmdLinesList[0], "SystemdUnitFile.Property() triggered Execute() called with unexpected cmd argument")
	assert.Equal(testCfg.ExitCodes, mockEnv.ExitCodesList[0], "SystemdUnitFile.Property() triggered util.Execute() called with unexpected exit codes")

}
