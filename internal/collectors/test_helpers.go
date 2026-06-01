package collectors

import (
	"fmt"
	"testing"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/mock"
)

const (
	// Dummy consts
	dummyBinPath = "/usr/local/bin/dummy"
)

type CommandVersionCall struct {
	output string
	count  int
	call   *mock.Call
}

func NewCommandVersionCall(output string) *CommandVersionCall {
	return &CommandVersionCall{
		output: output,
		count:  0,
		call:   nil,
	}
}

func (cvc *CommandVersionCall) UpdateExpectedCallCount() {
	// do nothing of call object not instantiated
	if cvc.call == nil {
		return
	}

	// update expected call count handling to reflect specified count
	if cvc.count <= 0 {
		// default to letting tests pass if no call count specified
		cvc.call.Maybe()
	} else if cvc.count > cvc.call.Repeatability {
		// update expected call count to reflect target count; the .Times()
		// value is additive so calculate necessary increment
		cvc.call.Times(cvc.count - cvc.call.Repeatability)
	}
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

	// command --version call management
	commandVersionCalls map[string]*CommandVersionCall

	// mockExecutor integration
	mockExecutor         util.MockExecutor
	mockExecutorTeardown func()
}

func NewSystemdTestEnv(t *testing.T) *systemdTestEnv {
	testEnv := new(systemdTestEnv)
	testEnv.Init(util.MOCK_SYSTEMD_SLE15SP3_VERSION)
	testEnv.Setup(t)
	return testEnv
}

func (env *systemdTestEnv) Init(systemVersion string) {
	env.units = []*util.SystemdUnit{}
	env.unitFiles = []*util.SystemdUnitFile{}
	env.execStarts = map[string]*util.SystemdServiceExecStart{}
	env.unitFileStates = map[util.SystemdUnitName]string{}
	env.commandVersionCalls = map[string]*CommandVersionCall{}
	env.SetSystemdVersion(systemVersion)
	env.mockExecutor = *util.NewMockExecutor()
}

func (env *systemdTestEnv) Setup(t *testing.T) {
	env.mockExecutorTeardown = env.mockExecutor.Setup(t)
}

func (env *systemdTestEnv) Cleanup() {
	//util.Execute = env.savedUtilExecute
	if env.mockExecutorTeardown != nil {
		env.mockExecutorTeardown()
	}
}

func (env *systemdTestEnv) SetSystemdVersion(systemdVersion string) {
	env.systemdVersion = systemdVersion
}

func (env *systemdTestEnv) AddCommandVersion(cmdPath string, version string) {
	env.commandVersionCalls[cmdPath] = NewCommandVersionCall(version)
}

func (env *systemdTestEnv) SetCommandVersionCallCount(cmdPath string, count int) {
	// set exepected call count
	cmdVer := env.commandVersionCalls[cmdPath]
	cmdVer.count = count

	// trigger update of associated call object's expected call count
	cmdVer.UpdateExpectedCallCount()
}

func (env *systemdTestEnv) RegisterCommandVersion(cmdPath string) {
	// lookup previously created command version management object
	cmdVer := env.commandVersionCalls[cmdPath]

	// register return values for the expected command line with the
	// expectation that it might be called
	cmdVer.call = env.mockExecutor.OnExecuteReturn(
		[]string{cmdPath, "--version"},
		[]int{0},
		[]byte(cmdVer.output),
		env.executeError,
	)

	// trigger update of associated call object's expected call count
	cmdVer.UpdateExpectedCallCount()
}

func determineUnitStates(svcState string) (activeState, subState string) {
	// determine unit state settings
	switch svcState {
	case "disabled":
		activeState = "inactive"
		subState = "dead"
	case "enabled":
		activeState = "active"
		subState = "running"
	default:
		activeState = "unknown"
		subState = "unknown"
	}

	return
}

func (env *systemdTestEnv) AddServiceUnit(svcName, svcDesc, svcState, cmdPath, versionOutput string) {
	// derived values
	unitName := util.SystemdUnitName(svcName)
	objectPath := fmt.Sprintf("/org/freedesktop/systemd1/unit/%s", svcName)
	activeState, subState := determineUnitStates(svcState)

	// create a SystemdUnit
	unit := util.NewSystemdUnit(
		unitName,
		svcDesc,
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
		cmdPath,
		[]string{
			cmdPath,
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
	env.AddCommandVersion(cmdPath, versionOutput)
}

func (env *systemdTestEnv) AddDummyUnit(svcName, svcState string) {
	env.AddServiceUnit(svcName, "dummy service", svcState, dummyBinPath, "dummy version")
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

	// setup handling for command --version calls
	for cmdPath := range env.commandVersionCalls {
		env.RegisterCommandVersion(cmdPath)
	}

	return mockClient
}
