package collectors

import (
	"fmt"
	"strings"
	"testing"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestPacemakerActiveNoPacemaker(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	res, err := isPacemakerRunning(systemdClient)

	// check returned values
	assert.NoError(err, "isPacemakerRunning() returned unexpected error")
	assert.Equal(false, res, "isPacemakerRunning() returned unexpected result")
}

func TestPacemakerActiveRunning(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.AddK8sUnit("dummy1.service", "dummy1 service", "server", "enabled")
	testEnv.AddK8sUnit("pacemaker.service", "Pacemaker High Availability Cluster Manager", "server", "enabled")
	testEnv.AddK8sUnit("dummy2.service", "dummy2 service", "server", "enabled")

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	res, err := isPacemakerRunning(systemdClient)

	// check returned values
	assert.NoError(err, "isPacemakerRunning() returned unexpected error")
	assert.Equal(true, res, "isPacemakerRunning() returned unexpected result")
}

func TestPacemakerActiveRunningNoPattern(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.AddK8sUnit("dummy1.service", "dummy1 service", "server", "enabled")
	testEnv.AddK8sUnit("pacemaker.service", "Pacemaker High Availability Cluster Manager", "server", "enabled")
	testEnv.AddK8sUnit("dummy2.service", "dummy2 service", "server", "enabled")
	// test using ListUnits
	testEnv.listUnitsByPatternsError = util.SystemdMethodNotAvailable

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	res, err := isPacemakerRunning(systemdClient)

	// check returned values
	assert.NoError(err, "isPacemakerRunning() returned unexpected error")
	assert.Equal(true, res, "isPacemakerRunning() returned unexpected result")
}

func TestPacemakerActiveNotRunning(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.AddK8sUnit("pacemaker.service", "Pacemaker High Availability Cluster Manager", "server", "disabled")

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	res, err := isPacemakerRunning(systemdClient)

	// check returned values
	assert.NoError(err, "isPacemakerRunning() returned unexpected error")
	assert.Equal(false, res, "isPacemakerRunning() returned unexpected result")
}

func TestPacemakerActiveErrorPattern(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.listUnitsByPatternsError = fmt.Errorf("test error")

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	res, err := isPacemakerRunning(systemdClient)

	// check returned values
	assert.Error(err, "isPacemakerRunning() failed to returned expected error")
	assert.Equal(false, res, "isPacemakerRunning() returned unexpected result")
}

func TestPacemakerActiveErrorUnits(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.listUnitsByPatternsError = util.SystemdMethodNotAvailable
	testEnv.listUnitsError = fmt.Errorf("test error")

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	res, err := isPacemakerRunning(systemdClient)

	// check returned values
	assert.Error(err, "isPacemakerRunning() failed to returned expected error")
	assert.Equal(false, res, "isPacemakerRunning() returned unexpected result")
}

func TestGetPacemakerVersion(t *testing.T) {
	// general setup
	assert := assert.New(t)
	var tests = []struct {
		input  string
		result string
	}{
		{"Pacemaker 2.1.10+20250718.fdf796ebc8-150700.3.3.1", "2.1.10+20250718.fdf796ebc8-150700.3.3.1"},
		{"Pacemaker", ""},
		{"Test Data", ""},
	}

	for _, v := range tests {
		// setup execute for pacemakerd --version
		mockUtilExecute(v.input, nil)
		// run test case
		res, err := getPacemakerVersion()

		// check returned values
		assert.NoError(err, "getPacemakerVersion() returned expected error")
		assert.Equal(v.result, res, "getPacemakerVersion() returned unexpected result")
	}
}

func TestGetPacemakerVersionExecuteError(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// setup execute for pacemakerd --version
	mockUtilExecute("", fmt.Errorf("forced error"))

	// run test case
	res, err := getPacemakerVersion()

	// check returned values
	assert.Error(err, "isSystemHA() failed to returned unexpected error")
	assert.ErrorContains(err, "forced error")
	assert.Equal("", res, "getPacemakerVersionNoInput() returned unexpected result")
}

func TestGetPacemakerVersionScannerError(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// we want to force a scanner error, which requires
	// execute returning a line > 64kb
	// from execute for pacemakerd --version
	mockUtilExecute(strings.Repeat("x", 65*1024), nil)

	// run test case
	res, err := getPacemakerVersion()

	// check returned values
	assert.Error(err, "getPacemakerVersioScannerErrorn() failed to returned expected error")
	assert.ErrorContains(err, "token too long")
	assert.Equal("", res, "getPacemakerVersionScannerError() returned unexpected result")
}

func TestIsPacemakerActiveNoPacemaker(t *testing.T) {
	// general setup
	assert := assert.New(t)
	expectedRes := Result{}

	// test settings
	testEnv := NewSystemdTestEnv()

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	res, err := isSystemHA(systemdClient)

	// check returned values
	assert.NoError(err, "isSystemHA() returned unexpected error")
	assert.Equal(expectedRes, res, "isSystemHA() returned unexpected result")
}

func TestIsPacemakerActiveRunningError(t *testing.T) {
	// general setup
	expectedRes := Result{}
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.AddK8sUnit("pacemaker.service", "Pacemaker High Availability Cluster Manager", "server", "enabled")

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// setup execute for pacemakerd --version
	mockUtilExecute("", fmt.Errorf("forced error"))

	// run test case
	res, err := isSystemHA(systemdClient)

	// check returned values
	assert.Error(err, "isSystemHA() failed to returned unexpected error")
	assert.ErrorContains(err, "forced error")
	assert.Equal(expectedRes, res, "isSystemHA() returned unexpected result")
}

func TestIsPacemakerActiveRunningSuccess(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv()
	testEnv.AddK8sUnit("pacemaker.service", "Pacemaker High Availability Cluster Manager", "server", "enabled")

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// setup execute for pacemakerd --version
	pacemakerVersion := "2.1.10+20250718.fdf796ebc8-150700.3.3.1"
	paceMakerName := "Pacemaker"
	expectedRes := Result{"ha_active": pacemakerVersion}

	// setup execute for pacemakerd --version
	mockUtilExecute(strings.Join([]string{paceMakerName, pacemakerVersion}, " "), nil)

	// run test case
	res, err := isSystemHA(systemdClient)

	// check returned values
	assert.NoError(err, "isSystemHA() returned unexpected error")
	assert.Equal(expectedRes, res, "isSystemHA() returned unexpected result")
}
