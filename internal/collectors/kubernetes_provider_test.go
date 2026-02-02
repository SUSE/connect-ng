package collectors

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// RKE2 consts
	rke2BinPath       = "/opt/rke2/bin/rke2"
	rke2ListPattern   = "rke2-*"
	rke2Version       = "v1.33.6+rke2r1"
	rke2VersionHash   = "2c2298232b55a94bd16b059f893c76a950811489"
	rke2VersionGolang = "go version go1.24.9 X:boringcrypto"

	// K3s consts
	k3sBinPath       = "/usr/local/bin/k3s"
	k3sListPattern   = "k3s*"
	k3sVersion       = "v1.33.6+k3s1"
	k3sVersionHash   = "b5847677"
	k3sVersionGolang = "go version go1.24.9"

	// Dummy consts
	dummyBinPath = "/usr/local/bin/dummy"
)

var (
	testSvcMatcher = regexp.MustCompile(`^test\d+\.service$`)
)

func TestMatchers(t *testing.T) {
	testUnitFile := "test.unit"
	enabled := "enabled"
	testBinPath := "/test/binary/path"
	testRole := "testRole"
	testVersion := "testVersion"
	testVerHash := "deadbeef"

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
			matcher:       listUnitFilesOutPutMatcher,
			input:         fmt.Sprintf("%s %s %s", testUnitFile, enabled, enabled),
			fieldMatches:  [][]byte{[]byte(testUnitFile), []byte(enabled)},
			expectedCount: 3,
		},
		{
			name:          "list-unit-files output matcher for invalid output (missing preset)",
			matcher:       listUnitFilesOutPutMatcher,
			input:         fmt.Sprintf("%s %s %s", testUnitFile, enabled, enabled),
			fieldMatches:  [][]byte{[]byte(testUnitFile), []byte(enabled)},
			expectedCount: 3,
		},
		{
			name:          "list-unit-files output matcher for invalid output (missing state and preset)",
			matcher:       listUnitFilesOutPutMatcher,
			input:         testUnitFile,
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},
		{
			name:          "list-unit-files output matcher for invalid output (empty)",
			matcher:       listUnitFilesOutPutMatcher,
			input:         "",
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},

		//
		// ExecStart property matcher
		//
		{
			name:          "ExecStart property matcher for valid ExecStart (path and argv with binary and role)",
			matcher:       execStartPropertyMatcher,
			input:         fmt.Sprintf("ExecStart={ path=%s ; argv[]=%s %s extra args ; ignore_errors=no ; start_time=[n/a] ; stop_time=[n/a] ; pid=0 ; code=(null) ; status=0/0 }", testBinPath, testBinPath, testRole),
			fieldMatches:  [][]byte{[]byte(testBinPath), []byte(testRole)},
			expectedCount: 3,
		},
		{
			name:          "ExecStart property matcher for valid ExecStart (path and argv with binary and role, no extra fields)",
			matcher:       execStartPropertyMatcher,
			input:         fmt.Sprintf("ExecStart={ path=%s ; argv[]=%s %s }", testBinPath, testBinPath, testRole),
			fieldMatches:  [][]byte{[]byte(testBinPath), []byte(testRole)},
			expectedCount: 3,
		},
		{
			name:          "ExecStart property matcher for invalid ExecStart (missing role in argv)",
			matcher:       execStartPropertyMatcher,
			input:         fmt.Sprintf("ExecStart={ path=%s ; argv[]=%s ; ignore_errors=no ; start_time=[n/a] ; stop_time=[n/a] ; pid=0 ; code=(null) ; status=0/0 }", testBinPath, testBinPath),
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},
		{
			name:          "ExecStart property matcher for invalid ExecStart (missing path)",
			matcher:       execStartPropertyMatcher,
			input:         fmt.Sprintf("ExecStart={ argv[]=%s %s ; ignore_errors=no ; start_time=[n/a] ; stop_time=[n/a] ; pid=0 ; code=(null) ; status=0/0 }", testBinPath, testRole),
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},
		{
			name:          "ExecStart property matcher for invalid ExecStart (missing argv)",
			matcher:       execStartPropertyMatcher,
			input:         fmt.Sprintf("ExecStart={ path=%s ; ignore_errors=no ; start_time=[n/a] ; stop_time=[n/a] ; pid=0 ; code=(null) ; status=0/0 }", testBinPath),
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},
		{
			name:          "ExecStart property matcher for invalid ExecStart (missing path and argv)",
			matcher:       execStartPropertyMatcher,
			input:         "ExecStart={ ignore_errors=no ; start_time=[n/a] ; stop_time=[n/a] ; pid=0 ; code=(null) ; status=0/0 }",
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},
		{
			name:          "ExecStart property matcher for invalid ExecStart (empty)",
			matcher:       execStartPropertyMatcher,
			input:         "",
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},

		//
		// rke2 service name matcher
		//
		{
			name:          "rke2 service name matcher for valid rke2 service name (server)",
			matcher:       rke2ServiceNameMatcher,
			input:         fmt.Sprintf("%s-%s.service", RKE2_PROVIDER, KUBERNETES_SERVER),
			fieldMatches:  [][]byte{},
			expectedCount: 1,
		},
		{
			name:          "rke2 service name matcher for invalid rke2 timer name (server)",
			matcher:       rke2ServiceNameMatcher,
			input:         fmt.Sprintf("%s-%s.timer", RKE2_PROVIDER, KUBERNETES_SERVER),
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},
		{
			name:          "rke2 service name matcher for valid rke2 service name (agent)",
			matcher:       rke2ServiceNameMatcher,
			input:         fmt.Sprintf("%s-%s.service", RKE2_PROVIDER, KUBERNETES_AGENT),
			fieldMatches:  [][]byte{},
			expectedCount: 1,
		},
		{
			name:          "rke2 service name matcher for invalid rke2 service name (invalid)",
			matcher:       rke2ServiceNameMatcher,
			input:         fmt.Sprintf("%s-%s.service", RKE2_PROVIDER, "invalid"),
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},

		//
		// k3s service name matcher
		//
		{
			name:          "k3s service name matcher for valid k3s service name (server)",
			matcher:       k3sServiceNameMatcher,
			input:         fmt.Sprintf("%s.service", K3S_PROVIDER),
			fieldMatches:  [][]byte{},
			expectedCount: 1,
		},
		{
			name:          "k3s service name matcher for invalid k3s timer name (server)",
			matcher:       k3sServiceNameMatcher,
			input:         fmt.Sprintf("%s.timer", K3S_PROVIDER),
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},
		{
			name:          "k3s service name matcher for valid k3s service name (agent)",
			matcher:       k3sServiceNameMatcher,
			input:         fmt.Sprintf("%s-%s.service", K3S_PROVIDER, KUBERNETES_AGENT),
			fieldMatches:  [][]byte{},
			expectedCount: 1,
		},
		{
			name:          "k3s service name matcher for invalid k3s timer name (agent)",
			matcher:       k3sServiceNameMatcher,
			input:         fmt.Sprintf("%s-%s.timer", K3S_PROVIDER, KUBERNETES_AGENT),
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},
		{
			name:          "k3s service name matcher for valid k3s service name (custom)",
			matcher:       k3sServiceNameMatcher,
			input:         fmt.Sprintf("%s-%s.service", K3S_PROVIDER, "custom"),
			fieldMatches:  [][]byte{},
			expectedCount: 1,
		},
		{
			name:          "k3s service name matcher for invalid k3s service name (invalid)",
			matcher:       k3sServiceNameMatcher,
			input:         fmt.Sprintf("%s%s.service", K3S_PROVIDER, "invalid"),
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},

		//
		// version matcher
		//
		{
			name:          "version matcher for valid version output",
			matcher:       versionMatcher,
			input:         fmt.Sprintf("test version %s (%s)", testVersion, testVerHash),
			fieldMatches:  [][]byte{[]byte(testVersion), []byte(testVerHash)},
			expectedCount: 3,
		},
		{
			name:          "version matcher for invalid version output (invalid hash)",
			matcher:       versionMatcher,
			input:         fmt.Sprintf("test version %s (%s)", testVersion, testVersion),
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},
		{
			name:          "version matcher for invalid version output (missing hash)",
			matcher:       versionMatcher,
			input:         fmt.Sprintf("test version %s", testVersion),
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},
		{
			name:          "version matcher for invalid version output (empty)",
			matcher:       versionMatcher,
			input:         fmt.Sprintf("test version %s", testVersion),
			fieldMatches:  [][]byte{},
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			matches := tc.matcher.FindSubmatch([]byte(tc.input))
			assert.Len(matches, tc.expectedCount, "FindSubMatch(%s) for %q didn't return expected number of matches", tc.matcher, tc.input)
			if len(matches) == 0 { // skip matched field checks not none expected
				return
			}
			assert.Equal([]byte(tc.input), matches[0], "FindSubmatch(%s) for %q should have matched entire input string", tc.matcher, tc.input)
			assert.Equal(tc.fieldMatches, matches[1:], "FindSubmatch(%s) for %q should have matched expected fields", tc.matcher, tc.input)
		})
	}
}

func TestSystemctlCmdNormalOperation(t *testing.T) {
	assert := assert.New(t)

	// test output
	testOutput := "test output"
	testSubCmd := "testSubCmd"
	testCmdArgs := []string{
		"testArg1",
		"testArg2",
	}

	// mock util.Execute
	calledCmdLine := []string{}
	calledExitCodes := []int{}
	origUtilExecute := util.Execute
	util.Execute = func(cmd []string, exitCodes []int) ([]byte, error) {
		// store arguments for later checking
		calledCmdLine = cmd
		calledExitCodes = exitCodes

		// return test output
		output := []byte(testOutput)
		return output, nil
	}
	defer func() { util.Execute = origUtilExecute }()

	// construct expected output and call command line and exit codes
	expectedCmdLine := []string{}
	expectedCmdLine = append(expectedCmdLine, systemctlBaseCmd...)
	expectedCmdLine = append(expectedCmdLine, testSubCmd)
	expectedCmdLine = append(expectedCmdLine, testCmdArgs...)
	expectedExitCodes := []int{0}
	expectedOutput := []byte(testOutput)

	// run test case
	output, err := systemctlCmd(testSubCmd, testCmdArgs...)

	// check returned values
	assert.Equal(expectedOutput, output, "systemctlCmd() returned unexpected output")
	assert.NoError(err, "systemctlCmd() returned unexpected error")

	// check util.Execute() was called as expected
	assert.Equal(expectedCmdLine, calledCmdLine, "systemctlCmd() called util.Execute() with unexperect cmd argument")
	assert.Equal(expectedExitCodes, calledExitCodes, "systemctlCmd() called util.Execute() with unexpected exitCodes argument")
}

func TestSystemctlCmdFailedOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)
	testEnv := newTestK8sEnvironment(testK8sEnvOptions{})
	callEnv := newTestK8sUtilExecuteEnv()

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// test output
	testSubCmd := "failingSubCmd"

	// construct expected output and call command line and exit codes
	expectedCmdLine := append(append([]string{}, systemctlBaseCmd...), testSubCmd)

	// run test case
	output, err := systemctlCmd(testSubCmd)

	// check returned values
	assert.Nil(output, "systemctlCmd() returned unexpected output")
	assert.ErrorIs(err, testEnv.DefaultError, "systemctlCmd() returned unexpected error")

	// check util.Execute() was called as expected
	assert.Equal(expectedCmdLine, callEnv.CmdLinesList[0], "systemctlCmd() called util.Execute() with unexperect cmd argument")
	require.Equal(1, callEnv.NumCalls, "util.Execute() called unexpected number of times")

	// check all util.Execute() exit codes are as expected
	for i, exitCodes := range callEnv.ExitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestListMatchingUnitFilesOfTypeAndStateNormalOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)
	testEnv := newTestK8sEnvironment(
		testK8sEnvOptions{
			Rke2Svcs: []*testK8sProviderService{
				newK8sProviderService("rke2-server.service", "server", generateExecStart(rke2BinPath, "server")),
				newK8sProviderService("rke2-agent.service", "agent", generateExecStart(rke2BinPath, "agent")),
			},
		},
	)
	callEnv := newTestK8sUtilExecuteEnv()
	tkp := testEnv.Rke2

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// run test case
	unitFiles, err := listMatchingUnitFilesOfTypeAndState(rke2ListPattern, "service", "enabled", rke2ServiceNameMatcher)

	// check returned values
	assert.NotNil(unitFiles, "listMatchingUnitFilesOfTypeAndState() returned nil for unitFiles list")
	assert.NoError(err, "listMatchingUnitFilesOfTypeAndState() returned unexpected error")
	assert.Equal([]*UnitFile{
		NewUnitFile(tkp.Services[0].Name),
		NewUnitFile(tkp.Services[1].Name),
	}, unitFiles, "listMatchingUnitFilesOfTypeAndState() returned unexpected output")

	// check util.Execute() was called as expected
	assert.Equal(testEnv.Rke2.ListUnitFilesCmd.CmdLine, callEnv.CmdLinesList[0], "listMatchingUnitFilesOfTypeAndState() called util.Execute() with unexperect cmd argument")
	require.Equal(1, callEnv.NumCalls, "util.Execute() called unexpected number of times")

	// check all util.Execute() exit codes are as expected
	for i, exitCodes := range callEnv.ExitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestListMatchingUnitFilesOfTypeAndStateNoMatchesOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)
	testEnv := newTestK8sEnvironment(testK8sEnvOptions{})
	callEnv := newTestK8sUtilExecuteEnv()

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// run test case
	unitFiles, err := listMatchingUnitFilesOfTypeAndState(rke2ListPattern, "service", "enabled", rke2ServiceNameMatcher)

	// check returned values
	assert.NotNil(unitFiles, "listMatchingUnitFilesOfTypeAndState() returned nil for unitFiles list")
	assert.NoError(err, "listMatchingUnitFilesOfTypeAndState() returned unexpected error")
	assert.Empty(unitFiles, "listMatchingUnitFilesOfTypeAndState() returned unexpected output")

	// check util.Execute() was called as expected
	assert.Equal(testEnv.Rke2.ListUnitFilesCmd.CmdLine, callEnv.CmdLinesList[0], "listMatchingUnitFilesOfTypeAndState() triggered util.Execute() called with unexperect cmd argument")
	require.Equal(1, callEnv.NumCalls, "util.Execute() called unexpected number of times")

	// check all util.Execute() exit codes are as expected
	for i, exitCodes := range callEnv.ExitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestListMatchingUnitFilesOfTypeAndStateFailedOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)
	testEnv := newTestK8sEnvironment(testK8sEnvOptions{
		FailSystemctlSubcommands: []string{"list-unit-files"},
	})
	callEnv := newTestK8sUtilExecuteEnv()

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// run test case
	unitFiles, err := listMatchingUnitFilesOfTypeAndState(rke2ListPattern, "service", "enabled", rke2ServiceNameMatcher)

	// check returned values
	assert.Nil(unitFiles, "listMatchingUnitFilesOfTypeAndState() returned unexpected output")
	assert.ErrorIs(err, testEnv.DefaultError, "listMatchingUnitFilesOfTypeAndState() returned unexpected error")

	// check util.Execute() was called as expected
	assert.Equal(testEnv.Rke2.ListUnitFilesCmd.CmdLine, callEnv.CmdLinesList[0], "listMatchingUnitFilesOfTypeAndState() triggered util.Execute() called with unexperect cmd argument")
	require.Equal(1, callEnv.NumCalls, "util.Execute() called unexpected number of times")

	// check all util.Execute() exit codes are as expected
	for i, exitCodes := range callEnv.ExitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestUnitFilePropertyNormalOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)
	testEnv := newTestK8sEnvironment(testK8sEnvOptions{
		Rke2Svcs: []*testK8sProviderService{
			newK8sProviderService("rke2-server.service", "server", generateExecStart(rke2BinPath, "server")),
		},
	})
	callEnv := newTestK8sUtilExecuteEnv()
	tks := testEnv.Rke2.Services[0]

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// run test case
	uf := NewUnitFile(tks.Name)
	output, err := uf.Property("ExecStart")

	// check returned values
	assert.Equal(tks.PropertyExecStartCmd.Output, output, "UnitFile.Property() returned unexpected output")
	assert.NoError(err, "UnitFile.Property() returned unexpected error")

	// check util.Execute() was called as expected
	assert.Equal(tks.PropertyExecStartCmd.CmdLine, callEnv.CmdLinesList[0], "unitFile.Property() triggered util.Execute() called with unexperect cmd argument")
	require.Equal(1, callEnv.NumCalls, "util.Execute() called unexpected number of times")

	// check all util.Execute() exit codes are as expected
	for i, exitCodes := range callEnv.ExitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestUnitFilePropertyFailedOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)
	testEnv := newTestK8sEnvironment(testK8sEnvOptions{
		Rke2Svcs: []*testK8sProviderService{
			newK8sProviderService("rke2-server.service", "server", generateExecStart(rke2BinPath, "server")),
		},
	})
	testEnv.Rke2.Disable() // force ExecStart property retrieval to fail
	callEnv := newTestK8sUtilExecuteEnv()
	tks := testEnv.Rke2.Services[0]

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// run test case
	uf := NewUnitFile(tks.Name)
	output, err := uf.Property("ExecStart")

	// check returned values
	assert.Nil(output, "UnitFile.Property() returned unexpected output")
	assert.ErrorIs(err, testEnv.DefaultError, "UnitFile.Property() returned unexpected error")

	// check util.Execute() was called as expected
	assert.Equal(tks.PropertyExecStartCmd.CmdLine, callEnv.CmdLinesList[0], "unitFile.Property() triggered util.Execute() called with unexperect cmd argument")
	require.Equal(1, callEnv.NumCalls, "util.Execute() called unexpected number of times")

	// check all util.Execute() exit codes are as expected
	for i, exitCodes := range callEnv.ExitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestKubernetesServiceInitNormalOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)
	testEnv := newTestK8sEnvironment(testK8sEnvOptions{
		Rke2Svcs: []*testK8sProviderService{
			newK8sProviderService("rke2-server.service", "server", generateExecStart(rke2BinPath, "server")),
		},
	})
	callEnv := newTestK8sUtilExecuteEnv()
	tkp := testEnv.Rke2
	tks := tkp.Services[0]

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// run test case
	uf := NewUnitFile(tks.Name)
	ks, err := NewKubernetesService(uf)

	// check returned values
	assert.NotNil(ks, "KubernetesService.Init() returned nil")
	assert.Equal(ks.unitFile, uf, "KubernetesService.Init() returned unexpected unitFile")
	assert.NoError(err, "KubernetesService.Init() returned unexpected error")

	// check KubernetesService.Init() extracted binary and role correctly
	assert.Equal(tkp.Binary, ks.Binary, "KubernetesService.Init() extracted binary incorrectly")
	assert.Equal(tks.Role, ks.Role, "KubernetesService.Init() extracted role incorrectly")
	assert.Equal(tkp.K8sType, ks.Type, "KubernetesService.Init() extracted Type incorrectly")
	assert.Equal(tkp.Version, ks.Version, "KubernetesService.setVersion() extracted version incorrectly")
	assert.Equal(tkp.VersionHash, ks.VersionHash, "KubernetesService.setVersion() extracted version hash incorrectly")

	// check util.Execute() called expected number of times and skip remaining tests if not
	require.Equal(2, callEnv.NumCalls, "util.Execute() called unexpected number of times")

	// check util.Execute() was called as expected for UnitFile.Property()
	assert.Equal(tks.PropertyExecStartCmd.CmdLine, callEnv.CmdLinesList[0], "UnitFile.Property() triggered util.Execute() with unexperect cmd argument")

	// check util.Execute() was called as expected for KubernetesService.setVersion()
	assert.Equal(tkp.VersionCmd.CmdLine, callEnv.CmdLinesList[1], "KubernetesService.setVersion() triggered util.Execute() with unexperect cmd argument")

	// check all util.Execute() exit codes are as expected
	for i, exitCodes := range callEnv.ExitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestKubernetesServiceInitFailedExecStartPropertyOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)
	testEnv := newTestK8sEnvironment(testK8sEnvOptions{
		Rke2Svcs: []*testK8sProviderService{
			newK8sProviderService("rke2-server.service", "server", generateExecStart(rke2BinPath, "server")),
		},
	})
	testEnv.Rke2.Disable() // force ExecStart property retrieval to fail
	callEnv := newTestK8sUtilExecuteEnv()
	tkp := testEnv.Rke2
	tks := tkp.Services[0]

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// run test case
	uf := NewUnitFile(tks.Name)
	ks, err := NewKubernetesService(uf)

	// check returned values
	assert.Nil(ks, "KubernetesService.Init() returned non-nil object")
	assert.ErrorIs(err, testEnv.DefaultError, "KubernetesService.Init() returned unexpected error")

	// check util.Execute() called expected number of times and skip remaining tests if not
	require.Equal(1, callEnv.NumCalls, "util.Execute() called unexpected number of times")

	// check util.Execute() was called as expected
	assert.Equal(tks.PropertyExecStartCmd.CmdLine, callEnv.CmdLinesList[0], "KubernetesService.Init() triggered util.Execute() with unexperect cmd argument")

	// check all util.Execute() exit codes are as expected
	for i, exitCodes := range callEnv.ExitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestKubernetesServiceInitFailedSetVersionOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)
	testEnv := newTestK8sEnvironment(testK8sEnvOptions{
		Rke2Svcs: []*testK8sProviderService{
			newK8sProviderService("rke2-server.service", "server", generateExecStart(rke2BinPath, "server")),
		},
		FailCommands: []string{rke2BinPath},
	})
	callEnv := newTestK8sUtilExecuteEnv()
	tkp := testEnv.Rke2
	tks := tkp.Services[0]

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// run test case
	uf := NewUnitFile(tks.Name)
	ks, err := NewKubernetesService(uf)

	// check returned values
	assert.Nil(ks, "KubernetesService.Init() returned non-nil object")
	assert.ErrorIs(err, testEnv.DefaultError, "KubernetesService.Init() returned unexpected error")

	// check util.Execute() called expected number of times and skip remaining tests if not
	require.Equal(2, callEnv.NumCalls, "util.Execute() called unexpected number of times")

	// check util.Execute() was called as expected for UnitFile.Property()
	assert.Equal(tks.PropertyExecStartCmd.CmdLine, callEnv.CmdLinesList[0], "UnitFile.Property() triggered util.Execute() with unexperect cmd argument")

	// check util.Execute() was called as expected for KubernetesService.setVersion()
	assert.Equal(tkp.VersionCmd.CmdLine, callEnv.CmdLinesList[1], "KubernetesService.setVersion() triggered util.Execute() with unexperect cmd argument")

	// check all util.Execute() exit codes are as expected
	for i, exitCodes := range callEnv.ExitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestGetKubernetesServicesNormalOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)
	testEnv := newTestK8sEnvironment(
		testK8sEnvOptions{
			Rke2Svcs: []*testK8sProviderService{
				newK8sProviderService("rke2-server.service", "server", generateExecStart(rke2BinPath, "server")),
				newK8sProviderService("rke2-agent.service", "agent", generateExecStart(rke2BinPath, "agent")),
				dummyK8sServiceOption("rke2-dummy.service"), // invalid service name
			},
			K3sSvcs: []*testK8sProviderService{
				newK8sProviderService("k3s.service", "server", generateExecStart(k3sBinPath, "server")),
				newK8sProviderService("k3s-agent.service", "agent", generateExecStart(k3sBinPath, "agent")),
				newK8sProviderService("k3s-custom.service", "server", generateExecStart(k3sBinPath, "server")),
				dummyK8sServiceOption("k3s-dummy.service"), // valid service name, invalid type & role
				dummyK8sServiceOption("k3snodash.service"), // invalid service name
			},
		},
	)
	callEnv := newTestK8sUtilExecuteEnv()

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// run test case
	ksList, err := getKubernetesServices()

	// check returned values
	assert.NotNil(ksList, "getKubernetesServices() returned nil")
	assert.NoError(err, "getKubernetesServices() returned unexpected error")
	require.Len(ksList, 5, "getKubernetesServices() returned unexpected number of services")

	// check discovered services
	for i, tks := range []*testK8sService{
		testEnv.Rke2.Services[0],
		testEnv.Rke2.Services[1],
		testEnv.K3s.Services[0],
		testEnv.K3s.Services[1],
		testEnv.K3s.Services[2],
	} {
		assert.Equal(tks.Name, ksList[i].unitFile.Name, "getKubernetesServices() returned unexpected service name for service %d", i)
		assert.Equal(tks.Role, ksList[i].Role, "getKubernetesServices() returned unexpected role for service [%d]%s", i, ksList[i].unitFile.Name)
	}

	// check discovered service provider details
	for i, tkp := range []*testK8sProvider{
		testEnv.Rke2,
		testEnv.Rke2,
		testEnv.K3s,
		testEnv.K3s,
		testEnv.K3s,
	} {
		assert.Equal(tkp.Binary, ksList[i].Binary, "getKubernetesServices() returned unexpected binary for service [%d]%s", i, ksList[i].unitFile.Name)
		assert.Equal(tkp.K8sType, ksList[i].Type, "getKubernetesServices() returned unexpected type for service [%d]%s", i, ksList[i].unitFile.Name)
		assert.Equal(tkp.Version, ksList[i].Version, "getKubernetesServices() returned unexpected version for service [%d]%s", i, ksList[i].unitFile.Name)
		assert.Equal(tkp.VersionHash, ksList[i].VersionHash, "getKubernetesServices() returned unexpected version hash for service [%d]%s", i, ksList[i].unitFile.Name)
	}

	// check util.Execute() called expected number of times and skip remaining tests if not
	expectedNumCalls := 2 /* list-unit-files for each provider pattern */ +
		2 /* show ExecStart property for valid Rke2 service names */ +
		4 /* show ExecStart property for valid K3s service names */ +
		2 /* version for each valid Rke2 service provider */ +
		3 /* version for each valid K3s service provider */

	require.Equal(expectedNumCalls, callEnv.NumCalls, "util.Execute() called unexpected number of times")

	// check util.Execute() was called as expected for listMatchingUnitFilesOfTypeAndState()
	for _, i := range []struct {
		tkp      *testK8sProvider
		cmdIndex int
	}{
		{testEnv.Rke2, 0},
		{testEnv.K3s, 5},
	} {
		assert.Equal(
			i.tkp.ListUnitFilesCmd.CmdLine,
			callEnv.CmdLinesList[i.cmdIndex],
			"listMatchingUnitFilesOfTypeAndState() triggered util.Execute() with unexpected cmd argument for %s",
			i.tkp.SvcPattern,
		)
	}

	// check util.Execute() was called as expected for UnitFile.Property() for valid services
	for _, i := range []struct {
		tkp                *testK8sProvider
		svcIndex, cmdIndex int
	}{
		{testEnv.Rke2, 0, 1},
		{testEnv.Rke2, 1, 3},
		{testEnv.K3s, 0, 6},
		{testEnv.K3s, 1, 8},
		{testEnv.K3s, 2, 10},
		{testEnv.K3s, 3, 12},
	} {
		svc := i.tkp.Services[i.svcIndex]
		assert.Equal(
			svc.PropertyExecStartCmd.CmdLine,
			callEnv.CmdLinesList[i.cmdIndex],
			"UnitFile.Property() triggered util.Execute() with unexpected cmd argument for %s",
			svc.Name,
		)
	}

	// check util.Execute() was called as expected for KubernetesService.setVersion()
	for _, i := range []struct {
		tkp      *testK8sProvider
		cmdIndex int
	}{
		{testEnv.Rke2, 2},
		{testEnv.Rke2, 4},
		{testEnv.K3s, 7},
		{testEnv.K3s, 9},
		{testEnv.K3s, 11},
	} {
		assert.Equal(
			i.tkp.VersionCmd.CmdLine,
			callEnv.CmdLinesList[i.cmdIndex],
			"KubernetesService.setVersion() triggered util.Execute() with unexpected cmd %d argument for %s",
			i,
			i.tkp.Binary)
	}

	// check all util.Execute() exit codes are as expected
	for i, exitCodes := range callEnv.ExitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestGetKubernetesProviderNormalOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)
	testEnv := newTestK8sEnvironment(
		testK8sEnvOptions{
			Rke2Svcs: []*testK8sProviderService{
				newK8sProviderService("rke2-server.service", "server", generateExecStart(rke2BinPath, "server")),
			},
		},
	)
	callEnv := newTestK8sUtilExecuteEnv()
	tkp := testEnv.Rke2
	tks := tkp.Services[0]

	// expected kubernetes provider info
	expectedKpInfo := map[string]any{
		"type":    tkp.K8sType,
		"role":    tks.Role,
		"version": tkp.Version,
	}

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// run test case
	kpInfo, err := getKubernetesProvider()

	// check returned values
	assert.NotNil(kpInfo, "getKubernetesProvider() returned nil")
	assert.NoError(err, "getKubernetesProvider() returned unexpected error")
	assert.Len(kpInfo, 3, "getKubernetesProvider() returned map with unexpected number of fields")
	assert.Equal(expectedKpInfo, kpInfo, "getKubernetesProvider() returned unexpected provider info")

	// check util.Execute() called expected number of times and skip remaining tests if not
	expectedNumCalls := 2 /* list-unit-files for each provider pattern */ +
		1 /* show ExecStart property for found RKE2 service name */ +
		1 /* version for found RKE2 service provider */

	require.Equal(expectedNumCalls, callEnv.NumCalls, "util.Execute() called unexpected number of times")

	// check util.Execute() was called as expected for listMatchingUnitFilesOfTypeAndState()
	for _, i := range []struct {
		tkp      *testK8sProvider
		cmdIndex int
	}{
		{testEnv.Rke2, 0},
		{testEnv.K3s, 3},
	} {
		assert.Equal(
			i.tkp.ListUnitFilesCmd.CmdLine,
			callEnv.CmdLinesList[i.cmdIndex],
			"listMatchingUnitFilesOfTypeAndState() triggered util.Execute() with unexpected cmd argument for %s",
			i.tkp.SvcPattern,
		)
	}

	// check util.Execute() was called as expected for UnitFile.Property() for found service
	assert.Equal(
		tks.PropertyExecStartCmd.CmdLine,
		callEnv.CmdLinesList[1],
		"UnitFile.Property() triggered util.Execute() with unexpected cmd argument for %s",
		tks.Name,
	)

	// check util.Execute() was called as expected for KubernetesService.setVersion()
	assert.Equal(
		tkp.VersionCmd.CmdLine,
		callEnv.CmdLinesList[2],
		"KubernetesService.setVersion() triggered util.Execute() with unexpected cmd argument for %s",
		tkp.Binary,
	)

	// check all util.Execute() exit codes are as expected
	for i, exitCodes := range callEnv.ExitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestGetKubernetesProviderNoServicesOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)
	testEnv := newTestK8sEnvironment(testK8sEnvOptions{})
	callEnv := newTestK8sUtilExecuteEnv()

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// expected
	// run test case
	kpInfo, err := getKubernetesProvider()

	// check returned values
	assert.Nil(kpInfo, "getKubernetesProvider() should have returned nil for provider info")
	assert.NoError(err, "getKubernetesProvider() returned unexpected error")

	// check util.Execute() called expected number of times and skip remaining tests if not
	expectedNumCalls := 2 /* list-unit-files for each provider pattern */

	require.Equal(expectedNumCalls, callEnv.NumCalls, "util.Execute() called unexpected number of times")

	// check util.Execute() was called as expected for listMatchingUnitFilesOfTypeAndState()
	for _, i := range []struct {
		tkp      *testK8sProvider
		cmdIndex int
	}{
		{testEnv.Rke2, 0},
		{testEnv.K3s, 1},
	} {
		assert.Equal(
			i.tkp.ListUnitFilesCmd.CmdLine,
			callEnv.CmdLinesList[i.cmdIndex],
			"listMatchingUnitFilesOfTypeAndState() triggered util.Execute() with unexpected cmd argument for %s",
			i.tkp.SvcPattern,
		)
	}

	// check all util.Execute() exit codes are as expected
	for i, exitCodes := range callEnv.ExitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestGetKubernetesProviderTooManyServicesOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)
	testEnv := newTestK8sEnvironment(
		testK8sEnvOptions{
			Rke2Svcs: []*testK8sProviderService{
				newK8sProviderService("rke2-server.service", "server", generateExecStart(rke2BinPath, "server")),
				newK8sProviderService("rke2-agent.service", "agent", generateExecStart(rke2BinPath, "agent")),
			},
		},
	)
	callEnv := newTestK8sUtilExecuteEnv()

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// expected
	// run test case
	kpInfo, err := getKubernetesProvider()

	// check returned values
	assert.Nil(kpInfo, "getKubernetesProvider() should have returned nil for provider info")
	assert.Error(err, "getKubernetesProvider() should have returned an error")
	assert.ErrorContains(err, "multiple kubernetes providers enabled", "getKubernetesProvider() returned unexpected error")

	// check util.Execute() called expected number of times and skip remaining tests if not
	expectedNumCalls := 2 /* list-unit-files for each provider pattern */ +
		2 /* show ExecStart property for found RKE2 service name */ +
		2 /* version for found RKE2 service provider */

	require.Equal(expectedNumCalls, callEnv.NumCalls, "util.Execute() called unexpected number of times")

	// check util.Execute() was called as expected for listMatchingUnitFilesOfTypeAndState()
	for _, i := range []struct {
		tkp      *testK8sProvider
		cmdIndex int
	}{
		{testEnv.Rke2, 0},
		{testEnv.K3s, 5},
	} {
		assert.Equal(
			i.tkp.ListUnitFilesCmd.CmdLine,
			callEnv.CmdLinesList[i.cmdIndex],
			"listMatchingUnitFilesOfTypeAndState() triggered util.Execute() with unexpected cmd argument for %s",
			i.tkp.SvcPattern,
		)
	}

	// check util.Execute() was called as expected for UnitFile.Property() for found services
	for _, i := range []struct {
		tkp                *testK8sProvider
		svcIndex, cmdIndex int
	}{
		{testEnv.Rke2, 0, 1},
		{testEnv.Rke2, 1, 3},
	} {
		svc := i.tkp.Services[i.svcIndex]
		assert.Equal(
			svc.PropertyExecStartCmd.CmdLine,
			callEnv.CmdLinesList[i.cmdIndex],
			"UnitFile.Property() triggered util.Execute() with unexpected cmd argument for %s",
			svc.Name,
		)
	}

	// check util.Execute() was called as expected for KubernetesService.setVersion()
	for _, i := range []struct {
		tkp      *testK8sProvider
		cmdIndex int
	}{
		{testEnv.Rke2, 2},
		{testEnv.Rke2, 4},
	} {
		assert.Equal(
			i.tkp.VersionCmd.CmdLine,
			callEnv.CmdLinesList[i.cmdIndex],
			"KubernetesService.setVersion() triggered util.Execute() with unexpected cmd argument for %s",
			i.tkp.Binary,
		)
	}

	// check all util.Execute() exit codes are as expected
	for i, exitCodes := range callEnv.ExitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestGetKubernetesProviderFailedOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)
	testEnv := newTestK8sEnvironment(
		testK8sEnvOptions{
			FailSystemctlSubcommands: []string{"list-unit-files"},
		},
	)
	callEnv := newTestK8sUtilExecuteEnv()
	tkp := testEnv.Rke2

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// expected
	// run test case
	kpInfo, err := getKubernetesProvider()

	// check returned values
	assert.Nil(kpInfo, "getKubernetesProvider() should have returned nil for provider info")
	assert.Error(err, "getKubernetesProvider() should have returned an error")
	assert.ErrorIs(err, testEnv.DefaultError, "getKubernetesProvider() returned unexpected error")

	// check util.Execute() called expected number of times and skip remaining tests if not
	expectedNumCalls := 1 /* list-unit-files for each provider pattern */

	require.Equal(expectedNumCalls, callEnv.NumCalls, "util.Execute() called unexpected number of times")

	// check util.Execute() was called as expected for listMatchingUnitFilesOfTypeAndState()
	assert.Equal(
		tkp.ListUnitFilesCmd.CmdLine,
		callEnv.CmdLinesList[0],
		"listMatchingUnitFilesOfTypeAndState() triggered util.Execute() with unexpected cmd argument for %s",
		tkp.SvcPattern,
	)

	// check all util.Execute() exit codes are as expected
	for i, exitCodes := range callEnv.ExitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestK8sRunSingleProvider(t *testing.T) {
	// general setup
	assert := assert.New(t)
	testEnv := newTestK8sEnvironment(
		testK8sEnvOptions{
			Rke2Svcs: []*testK8sProviderService{
				newK8sProviderService("rke2-server.service", "server", generateExecStart(rke2BinPath, "server")),
			},
		},
	)
	callEnv := newTestK8sUtilExecuteEnv()
	tkp := testEnv.Rke2
	tks := tkp.Services[0]

	// expected kubernetes provider info
	expectedResult := Result{
		"kubernetes_provider": map[string]any{
			"type":    tkp.K8sType,
			"role":    tks.Role,
			"version": tkp.Version,
		},
	}

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// run test case
	k8s := new(K8S)
	result, err := k8s.run(ARCHITECTURE_X86_64)

	// check returned values
	assert.NotNil(result, "K8S.run() returned nil")
	assert.NoError(err, "K8S.run() returned unexpected error")
	assert.Equal(expectedResult, result, "K8S.run() returned unexpected provider info")
}

func TestK8sRunNoProviders(t *testing.T) {
	// general setup
	assert := assert.New(t)
	testEnv := newTestK8sEnvironment(testK8sEnvOptions{})
	callEnv := newTestK8sUtilExecuteEnv()

	// expected kubernetes provider info
	expectedResult := Result{}

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// run test case
	k8s := new(K8S)
	result, err := k8s.run(ARCHITECTURE_X86_64)

	// check returned values
	assert.NotNil(result, "K8S.run() returned nil")
	assert.NoError(err, "K8S.run() returned unexpected error")
	assert.Equal(expectedResult, result, "K8S.run() returned unexpected provider info")
}

func TestK8sRunTooManyProviders(t *testing.T) {
	// general setup
	assert := assert.New(t)
	testEnv := newTestK8sEnvironment(
		testK8sEnvOptions{
			Rke2Svcs: []*testK8sProviderService{
				newK8sProviderService("rke2-server.service", "server", generateExecStart(rke2BinPath, "server")),
			},
			K3sSvcs: []*testK8sProviderService{
				newK8sProviderService("k3s.service", "server", generateExecStart(k3sBinPath, "server")),
			},
		},
	)
	callEnv := newTestK8sUtilExecuteEnv()

	// expected kubernetes provider info
	expectedResult := Result{}

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// run test case
	k8s := new(K8S)
	result, err := k8s.run(ARCHITECTURE_X86_64)

	// check returned values
	assert.NotNil(result, "K8S.run() returned nil")
	assert.Error(err, "K8S.run() should have returned an error")
	assert.ErrorContains(err, "multiple kubernetes providers enabled", "K8S.run() returned unexpected error")
	assert.Equal(expectedResult, result, "K8S.run() returned unexpected provider info")
}

func TestK8sRunGetKubernetesProvidersFailed(t *testing.T) {
	// general setup
	assert := assert.New(t)
	testEnv := newTestK8sEnvironment(testK8sEnvOptions{FailSystemctlSubcommands: []string{"list-unit-files"}})
	callEnv := newTestK8sUtilExecuteEnv()

	// expected kubernetes provider info
	expectedResult := Result{}

	// setup util.Execute() mocking
	origUtilExecute := setupUtilExecuteFunc(testEnv, callEnv)
	defer func() { util.Execute = origUtilExecute }()

	// run test case
	k8s := new(K8S)
	result, err := k8s.run(ARCHITECTURE_X86_64)

	// check returned values
	assert.NotNil(result, "K8S.run() returned nil")
	assert.Error(err, "K8S.run() should have returned an error")
	assert.ErrorIs(err, testEnv.DefaultError, "K8S.run() error not the expected one")
	assert.Equal(expectedResult, result, "K8S.run() returned unexpected provider info")
}

//
// util.Execute() mocking
//

type testK8sExecCmd struct {
	CmdLine   []string
	ExitCodes []int
	Output    []byte
	Err       error
}

func newTestK8sListUnitFilesCmd(
	svcPattern string,
	output string,
	err error,
) *testK8sExecCmd {
	cmdLine := append([]string{}, systemctlBaseCmd...)
	cmdLine = append(cmdLine, "list-unit-files", "--type", "service", "--state", "enabled", svcPattern)

	return &testK8sExecCmd{
		CmdLine:   cmdLine,
		ExitCodes: []int{0},
		Output:    []byte(output),
		Err:       err,
	}
}

func newTestK8sExecStartPropertyCmd(
	unitFile string,
	output string,
	err error,
) *testK8sExecCmd {
	cmdLine := append([]string{}, systemctlBaseCmd...)
	cmdLine = append(cmdLine, "show", "--property", "ExecStart", unitFile)

	return &testK8sExecCmd{
		CmdLine:   cmdLine,
		ExitCodes: []int{0},
		Output:    []byte(output),
		Err:       err,
	}
}

func newTestK8sVersionCmd(binary string, output string, err error) *testK8sExecCmd {
	return &testK8sExecCmd{
		CmdLine:   []string{binary, "--version"},
		ExitCodes: []int{0},
		Output:    []byte(output),
		Err:       err,
	}
}

type testK8sService struct {
	Name                 string
	Role                 string
	PropertyExecStartCmd testK8sExecCmd
}

type testK8sProviderService struct {
	Name      string
	Role      string
	ExecStart string
}

func generateExecStart(binPath string, args ...string) string {
	return fmt.Sprintf(
		"ExecStart={ path=%s ; argv[]=%s %s ; ignore_errors=no ; start_time=[n/a] ; stop_time=[n/a] ; pid=0 ; code=(null) ; status=0/0 }\n",
		binPath,
		binPath,
		strings.Join(args, " "),
	)
}

func newK8sProviderService(name, role, execStart string) *testK8sProviderService {
	return &testK8sProviderService{
		Name:      name,
		Role:      role,
		ExecStart: execStart,
	}
}

func dummyK8sServiceOption(name string) *testK8sProviderService {
	return newK8sProviderService(
		name,
		"dummy",
		generateExecStart(dummyBinPath, "dummy"),
	)
}

type testK8sProvider struct {
	Enabled          bool
	Binary           string
	K8sType          string
	Version          string
	VersionHash      string
	SvcPattern       string
	VersionCmd       *testK8sExecCmd
	ListUnitFilesCmd *testK8sExecCmd
	Services         []*testK8sService
}

func generatek8sProviderVersionOutput(binary, version, hash, golang string) string {
	return fmt.Sprintf(
		"%s version %s (%s)\n%s\n",
		filepath.Base(binary),
		version,
		hash,
		golang,
	)
}

func newTestK8sProvider(
	binary string,
	svcPattern string,
	version string,
	versionHash string,
	golangVersion string,
	services []*testK8sProviderService,
) *testK8sProvider {
	// derived values
	k8sType := filepath.Base(binary)
	versionOutput := generatek8sProviderVersionOutput(k8sType, version, versionHash, golangVersion)
	versionCmd := newTestK8sVersionCmd(binary, versionOutput, nil)

	tkp := &testK8sProvider{
		Enabled:     len(services) > 0,
		Binary:      binary,
		K8sType:     k8sType,
		SvcPattern:  svcPattern,
		VersionCmd:  versionCmd,
		Version:     version,
		VersionHash: versionHash,
		Services:    []*testK8sService{},
	}

	for _, svc := range services {
		tkp.AddService(svc.Name, svc.Role, svc.ExecStart)
	}

	tkp.SetupListUnitFilesCmd()

	return tkp
}

func (tkp *testK8sProvider) AddService(
	name string,
	role string,
	execStart string,
) {
	tks := &testK8sService{
		Name:                 name,
		Role:                 role,
		PropertyExecStartCmd: *newTestK8sExecStartPropertyCmd(name, execStart, nil),
	}
	tkp.Services = append(tkp.Services, tks)
}

func (tkp *testK8sProvider) SetupListUnitFilesCmd() {
	configuredSvcEntries := []string{}
	for _, svc := range tkp.Services {
		configuredSvcEntries = append(configuredSvcEntries, fmt.Sprintf("%s enabled enabled", svc.Name))
	}
	listUnitFilesOutput := fmt.Sprintf(
		"%s\n",
		strings.Join(configuredSvcEntries, "\n"),
	)
	tkp.ListUnitFilesCmd = newTestK8sListUnitFilesCmd(tkp.SvcPattern, listUnitFilesOutput, nil)
}

func (tkp *testK8sProvider) Enable() {
	tkp.Enabled = true
}

func (tkp *testK8sProvider) Disable() {
	tkp.Enabled = false
}

type testK8sEnvironment struct {
	Rke2                     *testK8sProvider
	K3s                      *testK8sProvider
	FailCommands             []string
	FailSystemctlSubcommands []string
	DefaultError             error
}

type testK8sEnvOptions struct {
	Rke2Svcs                 []*testK8sProviderService
	K3sSvcs                  []*testK8sProviderService
	FailCommands             []string
	FailSystemctlSubcommands []string
}

func newTestK8sEnvironment(opts testK8sEnvOptions) *testK8sEnvironment {
	tke := &testK8sEnvironment{
		DefaultError: fmt.Errorf("test error"),
	}

	tke.Rke2 = newTestK8sProvider(
		rke2BinPath,
		rke2ListPattern,
		rke2Version,
		rke2VersionHash,
		rke2VersionGolang,
		opts.Rke2Svcs,
	)

	tke.K3s = newTestK8sProvider(
		k3sBinPath,
		k3sListPattern,
		k3sVersion,
		k3sVersionHash,
		k3sVersionGolang,
		opts.K3sSvcs,
	)

	// setup forced command and systemctl subcommand failures
	tke.FailCommands = opts.FailCommands
	tke.FailSystemctlSubcommands = opts.FailSystemctlSubcommands

	return tke
}

type testK8sUtilExecuteEnv struct {
	NumCalls      int
	CmdLinesList  [][]string
	ExitCodesList [][]int
}

func newTestK8sUtilExecuteEnv() *testK8sUtilExecuteEnv {
	tkuee := &testK8sUtilExecuteEnv{
		NumCalls: 0,
	}
	return tkuee
}

func (tkuee *testK8sUtilExecuteEnv) Reset() {
	tkuee.NumCalls = 0
	tkuee.CmdLinesList = [][]string{}
	tkuee.ExitCodesList = [][]int{}
}

type testK8sExecuteFunc func(cmd []string, validExitCodes []int) ([]byte, error)

func setupUtilExecuteFunc(
	testEnv *testK8sEnvironment,
	callEnv *testK8sUtilExecuteEnv,
) testK8sExecuteFunc {
	oldExecute := util.Execute
	util.Execute = func(cmd []string, exitCodes []int) ([]byte, error) {
		// increment call counter
		callEnv.NumCalls++

		// store arguments for later checking
		callEnv.CmdLinesList = append(callEnv.CmdLinesList, cmd)
		callEnv.ExitCodesList = append(callEnv.ExitCodesList, exitCodes)

		// return test output
		var output []byte = nil
		var err error = nil
		switch {
		// fail if specified command should be forced to fail
		case slices.Contains(testEnv.FailCommands, cmd[0]):
			// return configured default error
			err = testEnv.DefaultError
		case cmd[0] == systemctlBin:
			subcmd := cmd[len(systemctlBaseCmd)]
			switch {
			// fail if specified systemctl subcommand should be forced to fail
			case slices.Contains(testEnv.FailSystemctlSubcommands, subcmd):
				// return configured default error
				err = testEnv.DefaultError
			// match on specific systemctl subcommands
			case subcmd == "list-unit-files":
				switch cmd[len(cmd)-1] {
				case testEnv.Rke2.SvcPattern:
					if testEnv.Rke2.Enabled {
						output = []byte(testEnv.Rke2.ListUnitFilesCmd.Output)
					} else {
						output = []byte("")
					}
				case testEnv.K3s.SvcPattern:
					if testEnv.K3s.Enabled {
						output = []byte(testEnv.K3s.ListUnitFilesCmd.Output)
					} else {
						output = []byte("")
					}
				default:
					// if matching pattern not found set error to default error
					err = testEnv.DefaultError
				}
			case subcmd == "show":
				// match on specific properties
				switch cmd[len(cmd)-2] {
				case "ExecStart":
					// look for service with matching name
					svcName := cmd[len(cmd)-1]

					// create list of candidate services
					candidateSvcs := []*testK8sService{}
					if testEnv.Rke2.Enabled {
						candidateSvcs = append(candidateSvcs, testEnv.Rke2.Services...)
					}
					if testEnv.K3s.Enabled {
						candidateSvcs = append(candidateSvcs, testEnv.K3s.Services...)
					}

					// iterate over candidate services looking for matching service
					var svc *testK8sService
					for _, tkp := range candidateSvcs {
						if tkp.Name == svcName {
							svc = tkp
							break
						}
					}

					// if matching service use it's PropertyExecStartCmd output otherwise set error to default error
					if svc != nil {
						output = []byte(svc.PropertyExecStartCmd.Output)
					} else {
						err = testEnv.DefaultError
					}
				}
			default:
				// return configured default error if subcommand is not recognised
				err = testEnv.DefaultError
			}
		case cmd[0] == testEnv.Rke2.Binary:
			output = []byte(testEnv.Rke2.VersionCmd.Output)
		case cmd[0] == testEnv.K3s.Binary:
			output = []byte(testEnv.K3s.VersionCmd.Output)
		default:
			// if matching binary not found set error to default error
			err = testEnv.DefaultError
		}
		return output, err

	}

	return oldExecute
}
