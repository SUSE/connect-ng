package collectors

import (
	"fmt"
	"path/filepath"
	"regexp"
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

func TestMatchers(t *testing.T) {
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

func generatek8sProviderVersionOutput(binary, version, hash, golang string) string {
	return fmt.Sprintf(
		"%s version %s (%s)\n%s\n",
		filepath.Base(binary),
		version,
		hash,
		golang,
	)
}

// helper code for testing systemd client operations
type systemdTestEnv struct {
	// forced errors
	closeError                   error
	listUnitsError               error
	listUnitsByPatternsError     error
	listUnitFilesError           error
	listUnitFilesByPatternsError error
	getExecStartError            error
	getSystemdVersionError       error
	getUnitFileStateError        error
	executeError                 error

	// systemd client response objects
	units          []*util.SystemdUnit
	unitFiles      []*util.SystemdUnitFile
	execStarts     map[string]*util.SystemdServiceExecStart
	systemdVersion string
	unitFileStates map[util.SystemdUnitName]string
	versionOutputs map[string]string

	// saved util.Execute
	savedUtilExecute func(cmd []string, validExitCodes []int) ([]byte, error)

	// mock util.Execute call state tracking
	numCalls      int
	cmdLinesList  [][]string
	exitCodesList [][]int
}

func NewSystemdTestEnv() *systemdTestEnv {
	testEnv := new(systemdTestEnv)
	testEnv.Init(util.MOCK_SYSTEMD_SLE15SP3_VERSION)
	return testEnv
}

func (env *systemdTestEnv) Init(systemVersion string) {
	env.units = []*util.SystemdUnit{}
	env.unitFiles = []*util.SystemdUnitFile{}
	env.execStarts = map[string]*util.SystemdServiceExecStart{}
	env.unitFileStates = map[util.SystemdUnitName]string{}
	env.versionOutputs = map[string]string{}
	env.SetSystemdVersion(systemVersion)
}

func (env *systemdTestEnv) SetSystemdVersion(systemdVersion string) {
	env.systemdVersion = systemdVersion
}

func (env *systemdTestEnv) AddK8sUnit(svcName, k8sType, k8sRole, svcState string) {
	// derived values
	unitName := util.SystemdUnitName(svcName)
	objectPath := fmt.Sprintf("/org/freedesktop/systemd1/unit/%s", svcName)

	// determine unit state settings
	var activeState string
	var subState string
	switch svcState {
	case "enabled":
		activeState = "inactive"
		subState = "dead"
	case "disabled":
		activeState = "active"
		subState = "running"
	default:
		activeState = "unknown"
		subState = "unknown"
	}

	// determine version string
	var binPath string
	var versionOutput string
	switch k8sType {
	case "rke2":
		binPath = rke2BinPath
		versionOutput = generatek8sProviderVersionOutput(binPath, rke2Version, rke2VersionHash, rke2VersionGolang)
	case "k3s":
		binPath = k3sBinPath
		versionOutput = generatek8sProviderVersionOutput(binPath, k3sVersion, k3sVersionHash, k3sVersionGolang)
	default:
		binPath = dummyBinPath
		versionOutput = "unknown"
	}

	// create a SystemdUnit
	unit := util.NewSystemdUnit(
		unitName,
		fmt.Sprintf("%s %s service", k8sType, k8sRole),
		"loaded",
		activeState,
		subState,
		"",
		objectPath,
		0,
		"",
		"/",
	)

	// create a SystemdUnitFile
	unitFile := util.NewSystemdUnitFile(
		fmt.Sprintf("/etc/systemd/system/%s", unitName),
		svcState,
	)

	// create a SystemdServiceExecStart
	execStart := util.NewSystemdServiceExecStart(
		binPath,
		[]string{
			binPath,
			k8sRole,
		},
		false,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
	)

	// add unit entry
	env.units = append(env.units, unit)

	// add unitFile entry
	env.unitFiles = append(env.unitFiles, unitFile)

	// add execStart entry
	env.execStarts[objectPath] = execStart

	// add state entry
	env.unitFileStates[unitName] = svcState

	// add version output
	env.versionOutputs[binPath] = versionOutput
}

func (env *systemdTestEnv) NewClient() util.SystemdClient {
	mockClient := util.NewMockSystemdClient(env.systemdVersion)

	mockClient.CloseFunc = func() error {
		if env.closeError != nil {
			return env.closeError
		}

		return nil
	}

	mockClient.ListUnitsFunc = func() ([]*util.SystemdUnit, error) {
		if env.listUnitsError != nil {
			return nil, env.listUnitsError
		}

		return env.units, nil
	}

	mockClient.ListUnitsByPatternsFunc = func(patterns ...string) ([]*util.SystemdUnit, error) {
		if env.listUnitsByPatternsError != nil {
			return nil, env.listUnitsByPatternsError
		}

		return env.units, nil
	}

	mockClient.ListUnitFilesFunc = func() ([]*util.SystemdUnitFile, error) {
		if env.listUnitFilesError != nil {
			return nil, env.listUnitFilesError
		}

		return env.unitFiles, nil
	}

	mockClient.ListUnitFilesByPatternsFunc = func(patterns ...string) ([]*util.SystemdUnitFile, error) {
		if env.listUnitFilesByPatternsError != nil {
			return nil, env.listUnitFilesByPatternsError
		}

		return env.unitFiles, nil
	}

	mockClient.GetExecStartFunc = func(unitObjectPath string) ([]*util.SystemdServiceExecStart, error) {
		execStarts := []*util.SystemdServiceExecStart{}

		if env.getExecStartError != nil {
			return nil, env.getExecStartError
		}

		execStart, ok := env.execStarts[unitObjectPath]
		if ok {
			execStarts = append(execStarts, execStart)
		}

		return execStarts, nil
	}

	mockClient.GetSystemdVersionFunc = func() (string, error) {
		if env.getSystemdVersionError != nil {
			return "", env.getSystemdVersionError
		}

		return env.systemdVersion, nil
	}

	mockClient.GetUnitFileStateFunc = func(unitName util.SystemdUnitName) (string, error) {
		state := "unknown"

		if env.getUnitFileStateError != nil {
			return "", env.getUnitFileStateError
		}
		if unitState, ok := env.unitFileStates[unitName]; ok {
			state = unitState
		}

		return state, nil
	}

	env.savedUtilExecute = util.Execute
	util.Execute = func(cmd []string, validExitCodes []int) ([]byte, error) {
		env.numCalls++
		env.cmdLinesList = append(env.cmdLinesList, cmd)
		env.exitCodesList = append(env.exitCodesList, validExitCodes)

		var output []byte
		var err error

		// return immediately an error has been specified
		if env.executeError != nil {
			return nil, env.executeError
		}

		switch cmd[0] {
		case rke2BinPath:
			fallthrough
		case k3sBinPath:
			switch cmd[1] {
			case "--version":
				output = []byte(env.versionOutputs[cmd[0]])
			default:
				output = []byte("some version output")
			}
		default:
			output = []byte("some command output")
		}

		return output, err
	}

	return mockClient
}

func (env *systemdTestEnv) Cleanup() {
	util.Execute = env.savedUtilExecute
}

func TestKubernetesServiceInitNormalOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	k8sSvc := "rke2-server.service"
	k8sType := RKE2_PROVIDER
	k8sRole := "server"
	k8sVersion := rke2Version
	k8sVersionHash := rke2VersionHash
	versionCmdLine := []string{rke2BinPath, "--version"}
	testEnv.AddK8sUnit(k8sSvc, k8sType, k8sRole, "enabled")

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	uf := testEnv.units[0]
	ks, err := NewKubernetesService(systemdClient, uf)

	// check returned values
	assert.NotNil(ks, "KubernetesService.Init() returned nil")
	assert.NoError(err, "KubernetesService.Init() returned unexpected error")

	// check KubernetesService.Init() extracted binary and role correctly
	assert.Equal(k8sSvc, ks.Name, "KubernetesService.Init() extracted binary incorrectly")
	assert.Equal(k8sType, ks.Type, "KubernetesService.Init() extracted type incorrectly")
	assert.Equal(k8sRole, ks.Role, "KubernetesService.Init() extracted role incorrectly")
	assert.Equal(rke2BinPath, ks.Binary, "KubernetesService.Init() extracted binary incorrectly")
	assert.Equal(k8sVersion, ks.Version, "KubernetesService.setVersion() extracted version incorrectly")
	assert.Equal(k8sVersionHash, ks.VersionHash, "KubernetesService.setVersion() extracted version hash incorrectly")

	// check util.Execute() called expected number of times and skip remaining tests if not
	require.Equal(1, testEnv.numCalls, "util.Execute() called unexpected number of times")

	// check util.Execute() was called as expected for KubernetesService.setVersion()
	assert.Equal(versionCmdLine, testEnv.cmdLinesList[0], "UnitFile.Property() triggered util.Execute() with unexperect cmd argument")

	// check all util.Execute() permitted exit codes are as expected
	for i, exitCodes := range testEnv.exitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestKubernetesServiceInitFailedExecStartPropertyOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")
	testEnv.getExecStartError = fmt.Errorf("test error")

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	uf := testEnv.units[0]
	ks, err := NewKubernetesService(systemdClient, uf)

	// check returned values
	assert.Nil(ks, "KubernetesService.Init() returned non-nil object")
	assert.ErrorIs(err, testEnv.getExecStartError, "KubernetesService.Init() returned unexpected error")
}

func TestKubernetesServiceInitFailedSetVersionOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")
	testEnv.executeError = fmt.Errorf("test error")

	// expected values
	versionCmdLine := []string{rke2BinPath, "--version"}

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	uf := testEnv.units[0]
	ks, err := NewKubernetesService(systemdClient, uf)

	// check returned values
	assert.Nil(ks, "KubernetesService.Init() returned non-nil object")
	assert.ErrorIs(err, testEnv.executeError, "KubernetesService.Init() returned unexpected error")

	// check util.Execute() called expected number of times and skip remaining tests if not
	require.Equal(1, testEnv.numCalls, "util.Execute() called unexpected number of times")

	// check util.Execute() was called as expected for KubernetesService.setVersion()
	assert.Equal(versionCmdLine, testEnv.cmdLinesList[0], "KubernetesService.setVersion() triggered util.Execute() with unexperect cmd argument")

	// check all util.Execute() permitted exit codes are as expected
	for i, exitCodes := range testEnv.exitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestGetKubernetesServicesNormalOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("rke2-agent.service", RKE2_PROVIDER, "agent", "enabled")
	testEnv.AddK8sUnit("rke2-dummy.service", RKE2_PROVIDER, "dummy", "enabled")
	testEnv.AddK8sUnit("k3s.service", K3S_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("k3s-agent.service", K3S_PROVIDER, "agent", "enabled")
	testEnv.AddK8sUnit("k3s-custom.service", K3S_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("k3s-dummy.service", K3S_PROVIDER, "dummy", "enabled")
	testEnv.AddK8sUnit("k3snodash.service", K3S_PROVIDER, "nodash", "enabled")

	// expected values
	rke2VersionCmdLine := []string{rke2BinPath, "--version"}
	k3sVersionCmdLine := []string{k3sBinPath, "--version"}

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	ksList, err := getKubernetesServices(systemdClient)

	// check returned values and fail immediatedly if less than expected services found
	assert.NotNil(ksList, "getKubernetesServices() returned nil")
	assert.NoError(err, "getKubernetesServices() returned unexpected error")
	require.Len(ksList, 5, "getKubernetesServices() returned unexpected number of services")

	// expect util.Execute() to have been called 5 times to retrieve version info
	require.Equal(5, testEnv.numCalls, "util.Execute() called unexpected number of times")

	// check discovered services
	for i, tks := range []struct {
		svcName, k8sType, k8sRole, binPath, version, hash string
		versionCmdLine                                    []string
	}{
		{
			"rke2-server.service",
			RKE2_PROVIDER,
			"server",
			rke2BinPath,
			rke2Version,
			rke2VersionHash,
			rke2VersionCmdLine,
		},
		{
			"rke2-agent.service",
			RKE2_PROVIDER,
			"agent",
			rke2BinPath,
			rke2Version,
			rke2VersionHash,
			rke2VersionCmdLine,
		},
		{
			"k3s.service",
			K3S_PROVIDER,
			"server",
			k3sBinPath,
			k3sVersion,
			k3sVersionHash,
			k3sVersionCmdLine,
		},
		{
			"k3s-agent.service",
			K3S_PROVIDER,
			"agent",
			k3sBinPath,
			k3sVersion,
			k3sVersionHash,
			k3sVersionCmdLine,
		},
		{
			"k3s-custom.service",
			K3S_PROVIDER,
			"server",
			k3sBinPath,
			k3sVersion,
			k3sVersionHash,
			k3sVersionCmdLine,
		},
	} {
		assert.Equal(tks.svcName, ksList[i].Name, "getKubernetesServices() returned unexpected service name for service %d", i)
		assert.Equal(tks.k8sType, ksList[i].Type, "getKubernetesServices() returned unexpected type for service [%d]%s", i, ksList[i].Name)
		assert.Equal(tks.k8sRole, ksList[i].Role, "getKubernetesServices() returned unexpected role for service [%d]%s", i, ksList[i].Name)
		assert.Equal(tks.binPath, ksList[i].Binary, "getKubernetesServices() returned unexpected binary path for service [%d]%s", i, ksList[i].Name)
		assert.Equal(tks.version, ksList[i].Version, "getKubernetesServices() returned unexpected version for service [%d]%s", i, ksList[i].Name)
		assert.Equal(tks.hash, ksList[i].VersionHash, "getKubernetesServices() returned unexpected version hash for service [%d]%s", i, ksList[i].Name)
		assert.Equal(tks.versionCmdLine, testEnv.cmdLinesList[i], "util.Execute() called with unexpected version command line for service [%d]%s", i, ksList[i].Name)
	}

	// check all util.Execute() exit codes are as expected
	for i, exitCodes := range testEnv.exitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestGetKubernetesServicesFallbackOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("k3s.service", K3S_PROVIDER, "server", "enabled")
	// Add a non-matching service to ensure filtering works via unit.Name.Match() in fallback
	testEnv.AddK8sUnit("unrelated.service", "dummy", "dummy", "enabled")

	// force fallback by simulating ListUnitsByPatterns not available
	testEnv.listUnitsByPatternsError = util.SystemdMethodNotAvailable

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	ksList, err := getKubernetesServices(systemdClient)

	// check returned values
	assert.NotNil(ksList, "getKubernetesServices() returned nil")
	assert.NoError(err, "getKubernetesServices() returned unexpected error")
	require.Len(ksList, 2, "getKubernetesServices() returned unexpected number of services")

	// check discovered services
	assert.Equal("rke2-server.service", ksList[0].Name)
	assert.Equal("k3s.service", ksList[1].Name)
}

func TestGetKubernetesServicesGetUnitFileStateError(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	testEnv := NewSystemdTestEnv()
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")

	// force state fetch to fail
	testEnv.getUnitFileStateError = fmt.Errorf("test state error")

	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	ksList, err := getKubernetesServices(systemdClient)

	assert.NotNil(ksList, "getKubernetesServices() returned nil")
	assert.NoError(err, "getKubernetesServices() returned unexpected error")
	require.Len(ksList, 0, "getKubernetesServices() should skip units with state errors")
}

func TestGetKubernetesProviderNormalOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("rke2-agent.service", RKE2_PROVIDER, "agent", "disabled")

	// expeected values
	versionCmdLine := []string{rke2BinPath, "--version"}

	// expected kubernetes provider info
	expectedKpInfo := map[string]any{
		"type":    RKE2_PROVIDER,
		"role":    "server",
		"version": rke2Version,
	}

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	kpInfo, err := getKubernetesProviderData(systemdClient)

	// check returned values
	assert.NotNil(kpInfo, "getKubernetesProvider() returned nil")
	assert.NoError(err, "getKubernetesProvider() returned unexpected error")
	assert.Len(kpInfo, 3, "getKubernetesProvider() returned map with unexpected number of fields")
	assert.Equal(expectedKpInfo, kpInfo, "getKubernetesProvider() returned unexpected provider info")

	// check util.Execute() called expected number of times and skip remaining tests if not
	require.Equal(1, testEnv.numCalls, "util.Execute() called unexpected number of times")

	// check util.Execute() was called as expected for KubernetesService.setVersion()
	assert.Equal(versionCmdLine, testEnv.cmdLinesList[0], "KubernetesService.setVersion() triggered util.Execute() with unexperect cmd argument")

	// check all util.Execute() permitted exit codes are as expected
	for i, exitCodes := range testEnv.exitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestGetKubernetesProviderNoServicesOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "disabled")
	testEnv.AddK8sUnit("rke2-agent.service", RKE2_PROVIDER, "agent", "disabled")

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	kpInfo, err := getKubernetesProviderData(systemdClient)

	// check returned values
	assert.Nil(kpInfo, "getKubernetesProvider() should have returned nil for provider info")
	assert.NoError(err, "getKubernetesProvider() returned unexpected error")

	// check util.Execute() called expected number of times and skip remaining tests if not
	require.Equal(0, testEnv.numCalls, "util.Execute() called unexpected number of times")
}

func TestGetKubernetesProviderTooManyServicesOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("k3s.service", K3S_PROVIDER, "server", "enabled")

	// expeected values
	rke2VersionCmdLine := []string{rke2BinPath, "--version"}
	k3sVersionCmdLine := []string{k3sBinPath, "--version"}

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	kpInfo, err := getKubernetesProviderData(systemdClient)

	// check returned values
	assert.Nil(kpInfo, "getKubernetesProvider() should have returned nil for provider info")
	assert.ErrorIs(err, KubernetesMultipleProvidersEnabled, "getKubernetesProvider() returned unexpected error")

	// check util.Execute() called expected number of times and skip remaining tests if not
	require.Equal(2, testEnv.numCalls, "util.Execute() called unexpected number of times")

	// check util.Execute() was called as expected for KubernetesService.setVersion()
	assert.Equal(rke2VersionCmdLine, testEnv.cmdLinesList[0], "KubernetesService.setVersion() triggered util.Execute() with unexperect cmd argument")
	assert.Equal(k3sVersionCmdLine, testEnv.cmdLinesList[1], "KubernetesService.setVersion() triggered util.Execute() with unexperect cmd argument")

	// check all util.Execute() permitted exit codes are as expected
	for i, exitCodes := range testEnv.exitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestGetKubernetesProviderFailedOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.listUnitsByPatternsError = fmt.Errorf("test error")

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	kpInfo, err := getKubernetesProviderData(systemdClient)

	// check returned values
	assert.Nil(kpInfo, "getKubernetesProvider() should have returned nil for provider info")
	assert.ErrorIs(err, testEnv.listUnitsByPatternsError, "getKubernetesProvider() returned unexpected error")

	// check util.Execute() called expected number of times and skip remaining tests if not
	require.Equal(0, testEnv.numCalls, "util.Execute() called unexpected number of times")
}

func TestGenerateKubernetesProviderNormalOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("rke2-agent.service", RKE2_PROVIDER, "agent", "disabled")

	// expected values
	versionCmdLine := []string{rke2BinPath, "--version"}

	// expected result info
	expectedKpInfo := map[string]any{
		"type":    RKE2_PROVIDER,
		"role":    "server",
		"version": rke2Version,
	}
	expectedResult := Result{"kubernetes_provider": expectedKpInfo}

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	res, err := generateKubernetesProvider(systemdClient)

	// check returned values
	assert.NotNil(res, "generateKubernetesProvider() returned nil")
	assert.NoError(err, "generateKubernetesProvider() returned unexpected error")
	assert.Equal(expectedResult, res, "generateKubernetesProvider() returned unexpected result")

	// check util.Execute() called expected number of times and skip remaining tests if not
	require.Equal(1, testEnv.numCalls, "util.Execute() called unexpected number of times")

	// check util.Execute() was called as expected for KubernetesService.setVersion()
	assert.Equal(versionCmdLine, testEnv.cmdLinesList[0], "KubernetesService.setVersion() triggered util.Execute() with unexperect cmd argument")

	// check all util.Execute() permitted exit codes are as expected
	for i, exitCodes := range testEnv.exitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestGenerateKubernetesProviderLegacyNormalOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.SetSystemdVersion(util.MOCK_SYSTEMD_SLE12SP5_VERSION)
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("rke2-agent.service", RKE2_PROVIDER, "agent", "disabled")
	testEnv.AddK8sUnit("rke2-dummy.service", RKE2_PROVIDER, "dummy", "enabled")
	testEnv.AddK8sUnit("k3s.service", K3S_PROVIDER, "server", "disabled")
	testEnv.AddK8sUnit("k3s-agent.service", K3S_PROVIDER, "agent", "disabled")
	testEnv.AddK8sUnit("k3s-dummy.service", K3S_PROVIDER, "dummy", "disabled")
	testEnv.AddK8sUnit("k3snodash.service", K3S_PROVIDER, "dummy", "disabled")

	// expected values
	versionCmdLine := []string{rke2BinPath, "--version"}

	// expected result info
	expectedKpInfo := map[string]any{
		"type":    RKE2_PROVIDER,
		"role":    "server",
		"version": rke2Version,
	}
	expectedResult := Result{"kubernetes_provider": expectedKpInfo}

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	res, err := generateKubernetesProvider(systemdClient)

	// check returned values
	assert.NotNil(res, "generateKubernetesProvider() returned nil")
	assert.NoError(err, "generateKubernetesProvider() returned unexpected error")
	assert.Equal(expectedResult, res, "generateKubernetesProvider() returned unexpected result")

	// check util.Execute() called expected number of times and skip remaining tests if not
	require.Equal(1, testEnv.numCalls, "util.Execute() called unexpected number of times")

	// check util.Execute() was called as expected for KubernetesService.setVersion()
	assert.Equal(versionCmdLine, testEnv.cmdLinesList[0], "KubernetesService.setVersion() triggered util.Execute() with unexperect cmd argument")

	// check all util.Execute() permitted exit codes are as expected
	for i, exitCodes := range testEnv.exitCodesList {
		assert.Equal([]int{0}, exitCodes, "util.Execute() returned unexpected exitCodes for call %d", i)
	}
}

func TestGenerateKubernetesProviderNoServicesOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "disabled")
	testEnv.AddK8sUnit("rke2-agent.service", RKE2_PROVIDER, "agent", "disabled")

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	res, err := generateKubernetesProvider(systemdClient)

	// check returned values
	assert.Equal(Result{}, res, "generateKubernetesProvider() should have returned an empty result")
	assert.NoError(err, "generateKubernetesProvider() returned unexpected error")

	// check util.Execute() called expected number of times and skip remaining tests if not
	require.Equal(0, testEnv.numCalls, "util.Execute() called unexpected number of times")
}

func TestGenerateKubernetesProviderTooManyServicesOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("k3s.service", K3S_PROVIDER, "server", "enabled")

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	res, err := generateKubernetesProvider(systemdClient)

	// check returned values
	assert.Equal(Result{}, res, "generateKubernetesProvider() should have returned an empty result")
	assert.ErrorIs(err, KubernetesMultipleProvidersEnabled, "generateKubernetesProvider() returned unexpected error")

	// check util.Execute() called expected number of times and skip remaining tests if not
	require.Equal(2, testEnv.numCalls, "util.Execute() called unexpected number of times")
}
func TestGenerateKubernetesProviderFailedOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.listUnitsByPatternsError = fmt.Errorf("test error")

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	res, err := generateKubernetesProvider(systemdClient)

	// check returned values
	assert.Equal(Result{}, res, "generateKubernetesProvider() should have returned an empty result")
	assert.ErrorIs(err, testEnv.listUnitsByPatternsError, "generateKubernetesProvider() returned unexpected error")

	// check util.Execute() called expected number of times and skip remaining tests if not
	require.Equal(0, testEnv.numCalls, "util.Execute() called unexpected number of times")
}
