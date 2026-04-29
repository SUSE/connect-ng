package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDbusSystemdClient_Integration attempts to use the real D-Bus connection.
// It skips the test if the system bus is not reachable (e.g. in CI or non-Linux environments).
func TestDbusSystemdClient_Integration(t *testing.T) {
	client, err := NewDbusSystemdClient()
	if err != nil {
		t.Skipf("Skipping systemd integration test: %v", err)
	}
	defer client.Close()

	require.NotZero(t, client.version, "GetSystemdVersion() returned empty string during Init()")

	// Basic connectivity check: List units and unit files
	_, err = client.ListUnits()
	require.NoError(t, err, "ListUnits() failed")

	_, err = client.ListUnitFiles()
	require.NoError(t, err, "ListUnitFiles() failed")
}
