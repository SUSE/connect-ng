package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSystemdUnitsByPatternsMultiplePatterns(t *testing.T) {
	assert := assert.New(t)

	command := []string{
		"org.freedesktop.systemd1",
		"/org/freedesktop/systemd1",
		"org.freedesktop.systemd1.Manager",
		"ListUnitsByPatterns",
		"asas",
		"0",
		"3",
		"sshd.service",
		"cron.service",
		"cups.service",
	}
	mockBusctlExecute(t, command, "internal/util/systemd_multiple_services.json")

	patterns := []string{"sshd.service", "cron.service", "cups.service"}
	units, err := SystemdUnitsByPatterns(patterns)

	assert.NoError(err)

	names := []string{}
	for _, unit := range units {
		names = append(names, unit.Name)
	}

	assert.ElementsMatch(patterns, names)
}

func TestSystemdUnitsByPatternsNoPattern(t *testing.T) {
	assert := assert.New(t)

	command := []string{
		"org.freedesktop.systemd1",
		"/org/freedesktop/systemd1",
		"org.freedesktop.systemd1.Manager",
		"ListUnitsByPatterns",
		"asas",
		"0",
		"1",
		"DOES_NOT_EXIST.service",
	}
	mockBusctlExecute(t, command, "internal/util/systemd_no_patterns_match.json")

	patterns := []string{"DOES_NOT_EXIST.service"}
	units, err := SystemdUnitsByPatterns(patterns)

	assert.NoError(err)
	assert.Equal(0, len(units))
}

func TestSystemdUnitsByPatternsBusFailed(t *testing.T) {
	assert := assert.New(t)

	mockBusctlExecuteFailed(t)

	patterns := []string{"sshd.service", "cron.service", "cups.service"}
	units, err := SystemdUnitsByPatterns(patterns)

	assert.Empty(units)
	assert.ErrorContains(err, "busctl call failed")
}

var systemdUnitSample = SystemdUnit{
	Name:        "k3s-agent.service",
	LoadedState: "loaded",
	ActiveState: "running",
	Path:        "/org/freedesktop/systemd1/unit/k3s_2dagent_2eservice",
}

func TestSystemdServiceBinPath(t *testing.T) {
	assert := assert.New(t)

	command := []string{
		"org.freedesktop.systemd1",
		systemdUnitSample.Path,
		"org.freedesktop.DBus.Properties",
		"Get",
		"ss",
		"org.freedesktop.systemd1.Service",
		"ExecStart",
	}

	mockBusctlExecute(t, command, "internal/util/systemd_service_unit_path_response.json")

	binPath, err := SystemdServiceBinPath(systemdUnitSample)

	assert.NoError(err)
	assert.Equal(binPath, "/usr/local/bin/k3s")
}

func TestSystemdServiceBinPathBusFailed(t *testing.T) {
	assert := assert.New(t)

	mockBusctlExecuteFailed(t)

	binPath, err := SystemdServiceBinPath(systemdUnitSample)

	assert.Empty(binPath)
	assert.ErrorContains(err, "busctl call failed")
}

func TestSystemdServiceBinPathInvalidStructure(t *testing.T) {
	assert := assert.New(t)

	command := []string{
		"org.freedesktop.systemd1",
		systemdUnitSample.Path,
		"org.freedesktop.DBus.Properties",
		"Get",
		"ss",
		"org.freedesktop.systemd1.Service",
		"ExecStart",
	}

	mockBusctlExecute(t, command, "internal/util/systemd_service_unit_path_response_invalid.json")

	binPath, err := SystemdServiceBinPath(systemdUnitSample)

	assert.Empty(binPath)
	assert.ErrorContains(err, "json: cannot unmarshal string into Go struct")
}
