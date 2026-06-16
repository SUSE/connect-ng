package collectors

import (
	"fmt"
	"strings"
	"testing"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/assert"
)

const (
	pacemakerName          = "Pacemaker"
	pacemakerDesc          = pacemakerName + " " + "High Availability Cluster Manager"
	pacemakerVersion       = "2.1.10+20250718.fdf796ebc8-150700.3.3.1"
	pacemakerVersionOutput = pacemakerName + " " + pacemakerVersion
)

func TestPacemakerActiveNoPacemaker(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv(t)

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
	testEnv := NewSystemdTestEnv(t)
	testEnv.AddDummyUnit("dummy1.service", "enabled")
	testEnv.AddServiceUnit("pacemaker.service", pacemakerDesc, "enabled", pacemakerCmdPath, pacemakerVersionOutput)
	testEnv.AddDummyUnit("dummy2.service", "enabled")

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
	testEnv := NewSystemdTestEnv(t)
	testEnv.AddDummyUnit("dummy1.service", "enabled")
	testEnv.AddServiceUnit("pacemaker.service", pacemakerDesc, "enabled", pacemakerCmdPath, pacemakerVersionOutput)
	testEnv.AddDummyUnit("dummy2.service", "enabled")

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
	testEnv := NewSystemdTestEnv(t)
	testEnv.AddServiceUnit("pacemaker.service", pacemakerDesc, "disabled", pacemakerCmdPath, pacemakerVersionOutput)

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
	testEnv := NewSystemdTestEnv(t)
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
	testEnv := NewSystemdTestEnv(t)
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

	// define test cases
	var testCases = []struct {
		name   string
		input  string
		result string
	}{
		{
			"Valid Version",
			pacemakerName + " " + pacemakerVersion,
			pacemakerVersion,
		},
		{
			"Empty Version",
			pacemakerName,
			"",
		},
		{
			"Invalid Version",
			"Test Data",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// create an Execute() mocker
			mockExecutor := util.NewMockExecutor()
			teardown := mockExecutor.Setup(t)
			defer teardown()

			// setup execute for pacemakerd --version
			mockExecutor.OnExecuteReturn([]string{pacemakerCmdPath, "--version"}, []int{0}, []byte(tc.input), nil).Once()

			// run test case
			res, err := getPacemakerVersion()

			// check returned values
			assert.NoError(err, "getPacemakerVersion() returned expected error")
			assert.Equal(tc.result, res, "getPacemakerVersion() returned unexpected result")
		})
	}
}

func TestGetPacemakerVersionExecuteError(t *testing.T) {
	// general setup
	assert := assert.New(t)
	testErr := fmt.Errorf("test error")

	// setup execute for pacemakerd --version
	mockExecutor := util.NewMockExecutor()
	mockExecutor.OnExecuteReturn([]string{pacemakerCmdPath, "--version"}, []int{0}, []byte{}, testErr).Once()
	teardown := mockExecutor.Setup(t)
	defer teardown()

	// run test case
	res, err := getPacemakerVersion()

	// check returned values
	assert.Error(err, "isSystemHA() failed to returned unexpected error")
	assert.ErrorIs(err, testErr, "isSystemHA() error is not the expected error")
	assert.Equal("", res, "getPacemakerVersionNoInput() returned unexpected result")
}

func TestGetPacemakerVersionScannerError(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// we want to force a scanner error, which requires
	// execute returning a line > 64kb
	invalidOutput := strings.Repeat("x", 65*1024)

	// setup execute for pacemakerd --version
	mockExecutor := util.NewMockExecutor()
	mockExecutor.OnExecuteReturn([]string{pacemakerCmdPath, "--version"}, []int{0}, []byte(invalidOutput), nil).Once()
	teardown := mockExecutor.Setup(t)
	defer teardown()

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
	testEnv := NewSystemdTestEnv(t)

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
	testEnv := NewSystemdTestEnv(t)
	testEnv.AddServiceUnit("pacemaker.service", pacemakerDesc, "enabled", pacemakerCmdPath, pacemakerVersionOutput)
	testEnv.executeError = fmt.Errorf("test error")

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// run test case
	res, err := isSystemHA(systemdClient)

	// check returned values
	assert.Error(err, "isSystemHA() failed to returned unexpected error")
	assert.ErrorIs(err, testEnv.executeError, "isSystemHA() error is not the expected error")
	assert.Equal(expectedRes, res, "isSystemHA() returned unexpected result")
}

func TestIsPacemakerActiveRunningSuccess(t *testing.T) {
	// general setup
	assert := assert.New(t)

	// test settings
	testEnv := NewSystemdTestEnv(t)
	testEnv.AddServiceUnit("pacemaker.service", pacemakerDesc, "enabled", pacemakerCmdPath, pacemakerVersionOutput)

	// create a mock systemd client
	systemdClient := testEnv.NewClient()
	defer testEnv.Cleanup()

	// setup execute for pacemakerd --version
	//paceMakerName := "Pacemaker"
	expectedRes := Result{"ha_active": pacemakerVersion}

	// setup execute for pacemakerd --version
	//mockUtilExecute(strings.Join([]string{paceMakerName, pacemakerVersion}, " "), nil)

	// run test case
	res, err := isSystemHA(systemdClient)

	// check returned values
	assert.NoError(err, "isSystemHA() returned unexpected error")
	assert.Equal(expectedRes, res, "isSystemHA() returned unexpected result")
}
