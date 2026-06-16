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
	rke2Version       = "v1.33.6+rke2r1"
	rke2VersionHash   = "2c2298232b55a94bd16b059f893c76a950811489"
	rke2VersionGolang = "go version go1.24.9 X:boringcrypto"

	// K3s consts
	k3sBinPath       = "/usr/local/bin/k3s"
	k3sVersion       = "v1.33.6+k3s1"
	k3sVersionHash   = "b5847677"
	k3sVersionGolang = "go version go1.24.9"
)

func TestKubernetesProviderMatchers(t *testing.T) {
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

func TestKubernetesServiceInitNormalOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv(t)
	rke2Svc := "rke2-server.service"
	rke2Type := RKE2_PROVIDER
	rke2Role := "server"
	testEnv.AddK8sUnit(rke2Svc, rke2Type, rke2Role, "enabled")

	// set expected command --version call counts
	testEnv.SetCommandVersionCallCount(rke2BinPath, 1)

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
	assert.Equal(rke2Svc, ks.Name, "KubernetesService.Init() extracted binary incorrectly")
	assert.Equal(rke2Type, ks.Type, "KubernetesService.Init() extracted type incorrectly")
	assert.Equal(rke2Role, ks.Role, "KubernetesService.Init() extracted role incorrectly")
	assert.Equal(rke2BinPath, ks.Binary, "KubernetesService.Init() extracted binary incorrectly")
	assert.Equal(rke2Version, ks.Version, "KubernetesService.setVersion() extracted version incorrectly")
	assert.Equal(rke2VersionHash, ks.VersionHash, "KubernetesService.setVersion() extracted version hash incorrectly")
}

func TestKubernetesServiceInitFailedExecStartPropertyOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv(t)
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

	// test settings
	testEnv := NewSystemdTestEnv(t)
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")
	testEnv.executeError = fmt.Errorf("test error")

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	uf := testEnv.units[0]
	ks, err := NewKubernetesService(systemdClient, uf)

	// check returned values
	assert.Nil(ks, "KubernetesService.Init() returned non-nil object")
	assert.ErrorIs(err, testEnv.executeError, "KubernetesService.Init() returned unexpected error")
}

func TestGetKubernetesServicesNormalOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test settings
	testEnv := NewSystemdTestEnv(t)
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("rke2-agent.service", RKE2_PROVIDER, "agent", "enabled")
	testEnv.AddK8sUnit("rke2-dummy.service", RKE2_PROVIDER, "dummy", "enabled")
	testEnv.AddK8sUnit("k3s.service", K3S_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("k3s-agent.service", K3S_PROVIDER, "agent", "enabled")
	testEnv.AddK8sUnit("k3s-custom.service", K3S_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("k3s-dummy.service", K3S_PROVIDER, "dummy", "enabled")
	testEnv.AddK8sUnit("k3snodash.service", K3S_PROVIDER, "nodash", "enabled")

	// set expected command --version call counts
	testEnv.SetCommandVersionCallCount(rke2BinPath, 2)
	testEnv.SetCommandVersionCallCount(k3sBinPath, 3)

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
	}
}

func TestGetKubernetesServicesFallbackOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	require := require.New(t)

	// test settings
	testEnv := NewSystemdTestEnv(t)
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("k3s.service", K3S_PROVIDER, "server", "enabled")
	// Add a non-matching service to ensure filtering works via unit.Name.Match() in fallback
	testEnv.AddK8sUnit("unrelated.service", "dummy", "dummy", "enabled")

	// set expected command --version call counts
	testEnv.SetCommandVersionCallCount(rke2BinPath, 1)
	testEnv.SetCommandVersionCallCount(k3sBinPath, 1)

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

	testEnv := NewSystemdTestEnv(t)
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

	// test settings
	testEnv := NewSystemdTestEnv(t)
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("rke2-agent.service", RKE2_PROVIDER, "agent", "disabled")

	// set expected command --version call counts
	testEnv.SetCommandVersionCallCount(rke2BinPath, 1)

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
}

func TestGetKubernetesProviderNoServicesOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv(t)
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
}

func TestGetKubernetesProviderTooManyServicesOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)
	//require := require.New(t)

	// test settings
	testEnv := NewSystemdTestEnv(t)
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("k3s.service", K3S_PROVIDER, "server", "enabled")

	// set expected command --version call counts
	testEnv.SetCommandVersionCallCount(rke2BinPath, 1)
	testEnv.SetCommandVersionCallCount(k3sBinPath, 1)

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	kpInfo, err := getKubernetesProviderData(systemdClient)

	// check returned values
	assert.Nil(kpInfo, "getKubernetesProvider() should have returned nil for provider info")
	assert.ErrorIs(err, KubernetesMultipleProvidersEnabled, "getKubernetesProvider() returned unexpected error")
}

func TestGetKubernetesProviderFailedOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv(t)
	testEnv.listUnitsByPatternsError = fmt.Errorf("test error")

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	kpInfo, err := getKubernetesProviderData(systemdClient)

	// check returned values
	assert.Nil(kpInfo, "getKubernetesProvider() should have returned nil for provider info")
	assert.ErrorIs(err, testEnv.listUnitsByPatternsError, "getKubernetesProvider() returned unexpected error")
}

func TestGenerateKubernetesProviderNormalOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv(t)
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("rke2-agent.service", RKE2_PROVIDER, "agent", "disabled")

	// set expected command --version call counts
	testEnv.SetCommandVersionCallCount(rke2BinPath, 1)

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
}

func TestGenerateKubernetesProviderLegacyNormalOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv(t)
	testEnv.SetSystemdVersion(util.MOCK_SYSTEMD_SLE12SP5_VERSION)
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("rke2-agent.service", RKE2_PROVIDER, "agent", "disabled")
	testEnv.AddK8sUnit("rke2-dummy.service", RKE2_PROVIDER, "dummy", "enabled")
	testEnv.AddK8sUnit("k3s.service", K3S_PROVIDER, "server", "disabled")
	testEnv.AddK8sUnit("k3s-agent.service", K3S_PROVIDER, "agent", "disabled")
	testEnv.AddK8sUnit("k3s-dummy.service", K3S_PROVIDER, "dummy", "disabled")
	testEnv.AddK8sUnit("k3snodash.service", K3S_PROVIDER, "dummy", "disabled")

	// set expected command --version call counts
	testEnv.SetCommandVersionCallCount(rke2BinPath, 1)

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
}

func TestGenerateKubernetesProviderNoServicesOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv(t)
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
}

func TestGenerateKubernetesProviderTooManyServicesOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv(t)
	testEnv.AddK8sUnit("rke2-server.service", RKE2_PROVIDER, "server", "enabled")
	testEnv.AddK8sUnit("k3s.service", K3S_PROVIDER, "server", "enabled")

	// set expected command --version call counts
	testEnv.SetCommandVersionCallCount(rke2BinPath, 1)
	testEnv.SetCommandVersionCallCount(k3sBinPath, 1)

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	res, err := generateKubernetesProvider(systemdClient)

	// check returned values
	assert.Equal(Result{}, res, "generateKubernetesProvider() should have returned an empty result")
	assert.ErrorIs(err, KubernetesMultipleProvidersEnabled, "generateKubernetesProvider() returned unexpected error")
}
func TestGenerateKubernetesProviderFailedOperation(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv(t)
	testEnv.listUnitsByPatternsError = fmt.Errorf("test error")

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	res, err := generateKubernetesProvider(systemdClient)

	// check returned values
	assert.Equal(Result{}, res, "generateKubernetesProvider() should have returned an empty result")
	assert.ErrorIs(err, testEnv.listUnitsByPatternsError, "generateKubernetesProvider() returned unexpected error")
}

//
// kubernetes provider specific systemdTestEnv extensions
//

func generatek8sProviderVersionOutput(binary, version, hash, golang string) string {
	return fmt.Sprintf(
		"%s version %s (%s)\n%s\n",
		filepath.Base(binary),
		version,
		hash,
		golang,
	)
}

func determinek8sVersionInfo(k8sType string) (binPath, versionOutput string) {
	// determine version string
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

	return
}

func (env *systemdTestEnv) AddK8sUnit(svcName, k8sType, k8sRole, svcState string) {
	// derived values
	unitName := util.SystemdUnitName(svcName)
	objectPath := fmt.Sprintf("/org/freedesktop/systemd1/unit/%s", svcName)
	activeState, subState := determineUnitStates(svcState)
	binPath, versionOutput := determinek8sVersionInfo(k8sType)

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
	env.AddCommandVersion(binPath, versionOutput)
}
