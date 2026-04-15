package util

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// define a error to be returned by our failure handlers
var forcedError = errors.New("forced")

// TestMockSystemdClient verifies that the MockSystemdClient implementation
// correctly delegates to the injected function fields. This ensures that the
// mocking mechanism behaves as expected for other tests.
func TestMockSystemdClient(t *testing.T) {

	// verify each of the interface methods
	t.Run("Close", testMockSystemdClientClose)
	t.Run("ListUnits", testMockSystemdClientListUnits)
	t.Run("ListUnitsByPatterns", testMockSystemdClientListUnitsByPatterns)
	t.Run("ListUnitFiles", testMockSystemdClientListUnitFiles)
	t.Run("ListUnitFilesByPatterns", testMockSystemdClientListUnitFilesByPatterns)
	t.Run("GetExecStart", testMockSystemdClientGetExecStart)
	t.Run("GetSystemdVersion", testMockSystemdClientGetSystemdVersion)
	t.Run("GetUnitFileState", testMockSystemdClientGetUnitFileState)
}

func testMockSystemdClientClose(t *testing.T) {
	// test setup
	assert := assert.New(t)

	// define a mock systemd client with a custom Close handler
	m := NewMockSystemdClient("")
	m.CloseFunc = func() error {
		return forcedError
	}

	// verify that custom handler is triggered
	err := m.Close()
	assert.Error(err, "Close() should return an error")
	assert.ErrorIs(err, forcedError, "Close() should return the expected error")

	// verify default nil behavior
	mClean := NewMockSystemdClient("")
	err = mClean.Close()
	assert.NoError(err, "Close() should not return an error for default mock handler")
}

func testMockSystemdClientListUnits(t *testing.T) {
	// test setup
	assert := assert.New(t)

	// define a response for a mock.service
	mockUnits := []*SystemdUnit{
		NewSystemdUnit(
			"mock.service",
			"Mock service",
			"active",
			"active",
			"",
			"",
			"/org/freedesktop/systemd1/unit/mock_2eservice",
			0,
			"",
			"",
		),
	}

	// define a mock systemd client with a custom ListUnits handler
	m := NewMockSystemdClient("")
	m.ListUnitsFunc = func() ([]*SystemdUnit, error) {
		return mockUnits, nil
	}

	// verify success mocking
	units, err := m.ListUnits()
	assert.NoError(err, "ListUnits() shouldn't have returned an error")
	assert.Equal(mockUnits, units, "ListUnits() should return the expected units")

	// verify failure mocking
	m.ListUnitsFunc = func() ([]*SystemdUnit, error) {
		return nil, forcedError
	}
	units, err = m.ListUnits()
	assert.Nil(units, "ListUnits() should have returned nil for failure")
	assert.Error(err, "ListUnits() should have returned an error")
	assert.ErrorIs(err, forcedError, "ListUnits() should have returned the expected error")

	// verify default nil behavior
	mClean := NewMockSystemdClient("")
	units, err = mClean.ListUnits()
	assert.Nil(units, "ListUnits() should have returned nil for default mock handler")
	assert.NoError(err, "ListUnits() should not have returned an error for default mock handler")
}

func testMockSystemdClientListUnitsByPatterns(t *testing.T) {
	// test setup
	assert := assert.New(t)

	// define a response for a mock.service
	mockUnits := []*SystemdUnit{
		NewSystemdUnit(
			"mock.service",
			"Mock service",
			"active",
			"active",
			"",
			"",
			"/org/freedesktop/systemd1/unit/mock_2eservice",
			0,
			"",
			"",
		),
	}

	// define a mock systemd client with a custom ListUnitsByPatterns handler
	m := NewMockSystemdClient("")
	m.ListUnitsByPatternsFunc = func(patterns ...string) ([]*SystemdUnit, error) {
		if len(patterns) > 0 && patterns[0] == "fail" {
			return nil, forcedError
		}
		return mockUnits, nil
	}

	// verify success mocking
	units, err := m.ListUnitsByPatterns("mock*")
	assert.NoError(err, "ListUnitsByPatterns() shouldn't have returned an error")
	assert.Equal(mockUnits, units, "ListUnitsByPatterns() should return the expected units")

	// verify failure mocking
	units, err = m.ListUnitsByPatterns("fail")
	assert.Nil(units, "ListUnitsByPatterns() should have returned nil for failure")
	assert.Error(err, "ListUnitsByPatterns() should have returned an error")
	assert.ErrorIs(err, forcedError, "ListUnitsByPatterns() should have returned the expected error")

	// verify legacy not available behavior
	mLegacy := NewMockSystemdClient(MOCK_SYSTEMD_SLE12SP5_VERSION)
	units, err = mLegacy.ListUnitsByPatterns("mock*")
	assert.Nil(units, "ListUnitsByPatterns() should have returned nil for not available method error")
	assert.ErrorIs(err, SystemdMethodNotAvailable, "ListUnitsByPatterns() should have returned not available method error")

	// verify default nil behavior
	mClean := NewMockSystemdClient("")
	units, err = mClean.ListUnitsByPatterns("ok")
	assert.Nil(units, "ListUnitsByPatterns() should have returned nil for default mock handler")
	assert.NoError(err, "ListUnitsByPatterns() should not have returned an error for default mock handler")
}

func testMockSystemdClientListUnitFiles(t *testing.T) {
	// test setup
	assert := assert.New(t)

	// define a response for a mock.service
	mockUnitFiles := []*SystemdUnitFile{
		NewSystemdUnitFile(
			"mock_2eservice",
			"enabled",
		),
	}

	// define a mock systemd client with a custom ListUnitFiles handler
	m := NewMockSystemdClient("")
	m.ListUnitFilesFunc = func() ([]*SystemdUnitFile, error) {
		return mockUnitFiles, nil
	}

	// verify success mocking
	units, err := m.ListUnitFiles()
	assert.NoError(err, "ListUnitFiles() shouldn't have returned an error")
	assert.Equal(mockUnitFiles, units, "ListUnitFiles() should return the expected units")

	// verify failure mocking
	m.ListUnitFilesFunc = func() ([]*SystemdUnitFile, error) {
		return nil, forcedError
	}
	units, err = m.ListUnitFiles()
	assert.Nil(units, "ListUnitFiles() should have returned nil for failure")
	assert.Error(err, "ListUnitFiles() should have returned an error")
	assert.ErrorIs(err, forcedError, "ListUnitFiles() should have returned the expected error")

	// verify default nil behavior
	mClean := NewMockSystemdClient("")
	units, err = mClean.ListUnitFiles()
	assert.Nil(units, "ListUnitFiles() should have returned nil for default mock handler")
	assert.NoError(err, "ListUnitFiles() should not have returned an error for default mock handler")
}

func testMockSystemdClientListUnitFilesByPatterns(t *testing.T) {
	// test setup
	assert := assert.New(t)

	// define a response for a mock.service
	mockUnitFiles := []*SystemdUnitFile{
		NewSystemdUnitFile(
			"mock_2eservice",
			"enabled",
		),
	}

	// define a mock systemd client with a custom ListUnitsByPatterns handler
	m := NewMockSystemdClient("")
	m.ListUnitFilesByPatternsFunc = func(patterns ...string) ([]*SystemdUnitFile, error) {
		if len(patterns) > 0 && patterns[0] == "fail" {
			return nil, forcedError
		}
		return mockUnitFiles, nil
	}

	// verify success mocking
	units, err := m.ListUnitFilesByPatterns("mock*")
	assert.NoError(err, "ListUnitFilesByPatterns() shouldn't have returned an error")
	assert.Equal(mockUnitFiles, units, "ListUnitFilesByPatterns() should return the expected units")

	// verify failure mocking
	units, err = m.ListUnitFilesByPatterns("fail")
	assert.Nil(units, "ListUnitFilesByPatterns() should have returned nil for failure")
	assert.Error(err, "ListUnitFilesByPatterns() should have returned an error")
	assert.ErrorIs(err, forcedError, "ListUnitFilesByPatterns() should have returned the expected error")

	// verify legacy not available behavior
	mLegacy := NewMockSystemdClient(MOCK_SYSTEMD_SLE12SP5_VERSION)
	units, err = mLegacy.ListUnitFilesByPatterns("mock*")
	assert.Nil(units, "ListUnitFilesByPatterns() should have returned nil for not available method error")
	assert.ErrorIs(err, SystemdMethodNotAvailable, "ListUnitFilesByPatterns() should have returned not available method error")

	// verify default nil behavior
	mClean := NewMockSystemdClient("")
	units, err = mClean.ListUnitFilesByPatterns("ok")
	assert.Nil(units, "ListUnitFilesByPatterns() should have returned nil for default mock handler")
	assert.NoError(err, "ListUnitFilesByPatterns() should not have returned an error for default mock handler")
}

func testMockSystemdClientGetExecStart(t *testing.T) {
	// test setup
	assert := assert.New(t)

	// define a response for a mock.service
	mockExecStart := []*SystemdServiceExecStart{
		NewSystemdServiceExecStart(
			"/some/path",
			[]string{
				"/some/path",
				"arg1",
				"arg2",
			},
			true,
			0,
			0,
			0,
			0,
			0,
			0,
			0,
		),
	}

	// define a mock systemd client with a custom GetExecStart handler
	m := NewMockSystemdClient("")
	m.GetExecStartFunc = func(unitPath string) ([]*SystemdServiceExecStart, error) {
		if unitPath == "fail" {
			return nil, forcedError
		}

		return mockExecStart, nil
	}

	// verify success mocking
	execStarts, err := m.GetExecStart("/org/freedesktop/systemd1/unit/mock_2eservice")
	assert.Equal(mockExecStart, execStarts, "GetExecStart() should return the expected response")
	assert.NoError(err, "GetExecStart() shouldn't have returned an error")

	// verify failure mocking
	execStarts, err = m.GetExecStart("fail")
	assert.Nil(execStarts, "GetExecStart() should have returned nil for failure")
	assert.Error(err, "GetExecStart() should have returned an error")
	assert.ErrorIs(err, forcedError, "GetExecStart() should have returned the expected error")

	// verify default nil behavior
	mClean := NewMockSystemdClient("")
	execStarts, err = mClean.GetExecStart("/org/freedesktop/systemd1/unit/mock_2eservice")
	assert.Nil(execStarts, "GetExecStart() should have returned nil for default mock handler")
	assert.NoError(err, "GetExecStart() should not have returned an error for default mock handler")
}

func testMockSystemdClientGetSystemdVersion(t *testing.T) {
	// test setup
	assert := assert.New(t)

	mockVersion := "246"

	// define a mock systemd client with a custom GetSystemdVersion handler
	m := NewMockSystemdClient("")
	m.GetSystemdVersionFunc = func() (string, error) {
		return mockVersion, nil
	}

	// verify success mocking
	version, err := m.GetSystemdVersion()
	assert.Equal(mockVersion, version, "GetSystemdVersion() should return the expected response")
	assert.NoError(err, "GetSystemdVersion() shouldn't have returned an error")

	// verify failure mocking
	m.GetSystemdVersionFunc = func() (string, error) {
		return "", forcedError
	}
	version, err = m.GetSystemdVersion()
	assert.Empty(version, "GetSystemdVersion() should have returned an empty string for failure")
	assert.Error(err, "GetSystemdVersion() should have returned an error")
	assert.ErrorIs(err, forcedError, "GetSystemdVersion() should have returned the expected error")

	// verify default nil behavior
	mClean := NewMockSystemdClient("")
	version, err = mClean.GetSystemdVersion()
	assert.Empty(version, "GetSystemdVersion() should have returned an empty string for default mock handler")
	assert.NoError(err, "GetSystemdVersion() should not have returned an error for default mock handler")
}

func testMockSystemdClientGetUnitFileState(t *testing.T) {
	// test setup
	assert := assert.New(t)

	mockState := "enabled"

	// define a mock systemd client with a custom GetUnitFileState handler
	m := NewMockSystemdClient("")
	m.GetUnitFileStateFunc = func(unitName SystemdUnitName) (string, error) {
		if unitName == "fail" {
			return "", forcedError
		}
		return mockState, nil
	}

	// verify success mocking
	state, err := m.GetUnitFileState("/org/freedesktop/systemd1/unit/mock_2eservice")
	assert.Equal(mockState, state, "GetUnitFileState() should return the expected response")
	assert.NoError(err, "GetUnitFileState() shouldn't have returned an error")

	// verify success mocking
	state, err = m.GetUnitFileState("fail")
	assert.Empty(state, "GetUnitFileState() should have returned an empty string for failure")
	assert.Error(err, "GetUnitFileState() should have returned an error")
	assert.ErrorIs(err, forcedError, "GetUnitFileState() should have returned the expected error")

	// verify default nil behavior
	mClean := NewMockSystemdClient("")
	state, err = mClean.GetUnitFileState("/org/freedesktop/systemd1/unit/mock_2eservice")
	assert.Empty(state, "GetUnitFileState() should have returned an empty string for default mock handler")
	assert.NoError(err, "GetExecStart() should not have returned an error for default mock handler")
}
