package util

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test the parseSystemdVersion() helper
func TestParseSystemdVersion(t *testing.T) {
	// general test setup
	assert := assert.New(t)

	// Define test cases
	testCases := []struct {
		name            string
		version         string
		expectedVersion uint64
		expectedError   error
	}{
		{ // SLE 12 SP 5
			name:            "Valid simple version",
			version:         "228",
			expectedVersion: 228,
			expectedError:   nil,
		},
		{ // SLE 15 SP3
			name:            "Valid complex version",
			version:         "246.16+suse.312.gb540e1826d",
			expectedVersion: 246,
			expectedError:   nil,
		},
		{ // invalid number
			name:            "Invalid version number",
			version:         "-123",
			expectedVersion: 0,
			expectedError:   strconv.ErrSyntax,
		},
		{ // invalid value
			name:            "Invalid version",
			version:         "invalid",
			expectedVersion: 0,
			expectedError:   strconv.ErrSyntax,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			version, err := parseSystemdVersion(tt.version)
			if tt.expectedError != nil {
				assert.ErrorIs(err, tt.expectedError, "Unexpected parsing error when parsing version")
			} else {
				assert.NoError(err, "Version parsing should not have failed")
				assert.Equal(tt.expectedVersion, version, "Parsed version does not match expected value")
			}
		})
	}
}

// Test correct operation of the UnitType() routine associated with SystemdUnitName type
func TestSystemdUnitName(t *testing.T) {
	tests := []struct {
		input SystemdUnitName
		ext   string
	}{
		{"mock.service", "service"},
		{"mock.target", "target"},
		{"mock.timer", "timer"},
		{"mock", ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			assert.Equal(t, tt.ext, tt.input.UnitType())
		})
	}
}

// Test correct operation of the UnitName() routine associated with SystemdUnitName type
func TestSystemdUnitName_UnitName(t *testing.T) {
	tests := []struct {
		input    SystemdUnitName
		expected string
	}{
		{"mock.service", "mock.service"},
		{"/etc/systemd/system/mock.service", "mock.service"},
		{"mock", "mock"},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.UnitName())
		})
	}
}

// Test correct operation of the NameMatch() routine associated with SystemdUnitName type
func TestSystemdUnitName_Match(t *testing.T) {
	tests := []struct {
		name     string
		input    SystemdUnitName
		patterns []string
		match    bool
	}{
		{"match_prefix", "mock.service", []string{"mock.*"}, true},
		{"match_suffix", "mock.service", []string{"*.service"}, true},
		{"match_multiple", "mock.service", []string{"test.*", "mock.*"}, true},
		{"no_match", "mock.service", []string{"test.*", "*.target"}, false},
		{"match_with_path", "/etc/systemd/system/mock.service", []string{"mock.*"}, true},
		{"empty_patterns", "mock.service", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.match, tt.input.Match(tt.patterns...))
		})
	}
}

// Validate correct operation of constructors
func TestSystemdConstructors(t *testing.T) {
	t.Run("NewSystemdUnitFile", func(t *testing.T) {
		unit := NewSystemdUnitFile("test.service", "enabled")
		assert.Equal(t, SystemdUnitName("test.service"), unit.Name)
		assert.Equal(t, "enabled", unit.State)
	})

	t.Run("NewSystemdUnit", func(t *testing.T) {
		unit := NewSystemdUnit(
			"test.service",
			"Test Description",
			"loaded",
			"active",
			"running",
			"none",
			"/object/path/to/test_2eservice",
			123,
			"start",
			"/path/to/job",
		)

		assert.Equal(t, SystemdUnitName("test.service"), unit.Name)
		assert.Equal(t, "/object/path/to/test_2eservice", unit.ObjectPath)
		assert.Equal(t, "Test Description", unit.Description)
		assert.Equal(t, "loaded", unit.LoadState)
		assert.Equal(t, "active", unit.ActiveState)
		assert.Equal(t, "running", unit.SubState)
		assert.Equal(t, "none", unit.Following)
		assert.Equal(t, uint32(123), unit.JobId)
		assert.Equal(t, "start", unit.JobType)
		assert.Equal(t, "/path/to/job", unit.JobPath)
	})

	t.Run("NewSystemdServiceExecStart", func(t *testing.T) {
		execStart := NewSystemdServiceExecStart(
			"/bin/true",
			[]string{"/bin/true", "--version"},
			true,
			1000,
			2000,
			3000,
			4000,
			9999,
			0,
			1,
		)

		assert.Equal(t, "/bin/true", execStart.Path)
		assert.Equal(t, []string{"/bin/true", "--version"}, execStart.Args)
		assert.True(t, execStart.IgnoreErrors)
		assert.Equal(t, uint64(1000), execStart.StartTimestamp)
		assert.Equal(t, uint64(2000), execStart.StartTimestampMonotonic)
		assert.Equal(t, uint64(3000), execStart.ExitTimestamp)
		assert.Equal(t, uint64(4000), execStart.ExitTimestampMonotonic)
		assert.Equal(t, uint32(9999), execStart.Pid)
		assert.Equal(t, int32(0), execStart.ExitCode)
		assert.Equal(t, int32(1), execStart.ExitStatus)
	})
}
