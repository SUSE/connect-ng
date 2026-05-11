package features

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/SUSE/connect-ng/features/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getKubernetesProviderInfo(t *testing.T) map[string]any {
	cli := helpers.NewRunner(t, "suseconnect --info")
	cli.Run()
	require.Equal(t, 0, cli.ExitCode(), "suseconnect --info failed")

	hwinfo := helpers.AssertValidJSON[map[string]any](t, cli.Stdout())
	if kp, ok := hwinfo["kubernetes_provider"].(map[string]any); ok {
		return kp
	}
	return nil
}

var (
	// path under which simulated k8s provider scripts will be created
	simulatedK8sDir = "/usr/local/bin"

	// directory for simulated systemd services
	etcSystemdDir = "/etc/systemd/system"

	// k8s provider versions - must match values in embedded script
	rke2Version = "v1.33.6+rke2r1"
	k3sVersion  = "v1.33.6+k3s1"

	// environment variables used to manage running these tests
	k8sProviderTestsEnabled   = "K8S_PROVIDER_TESTS_ENABLED"
	k8sProviderTestsNoCleanup = "K8S_PROVIDER_TESTS_NO_CLEANUP"
)

func generateSimulatedK8sScript() string {
	return `#!/bin/bash

# ensure unassigned variables are treated as errors and fail immediately
set -eu

# determine the script name, which is the k8s provider to simulate
k8s="$(basename "${BASH_SOURCE[0]}")"

declare -A is_simulated
is_simulated[k3s]=true
is_simulated[rke2]=true

if [[ "${is_simulated[${k8s}]:-false}" != "true" ]]; then
    echo "ERROR: k8s provider not simulated: '${k8s}'"
    exit 1
fi

declare -A k8s_version_info
# simulated k3s version string - keep in sync with values in test file
k8s_version_info[k3s]="k3s version ` + k3sVersion + ` (b5847677)
go version go1.24.9"
# simulated rke2 version string
k8s_version_info[rke2]="rke2 version ` + rke2Version + ` (2c2298232b55a94bd16b059f893c76a950811489)
go version go1.24.9 X:boringcrypto"

# simulate a running service by sleeping forever until
# terminated by a signal
simulate_service_running()
{
    local role="${1-server}"

    # trap signals to exit cleanly when systemd stops the service
    trap "echo Exiting...; exit 0" SIGINT SIGTERM

    # sleep forever so systemd maintains an 'active' state
    echo "Simulated ${k8s} sleeping forever simulating role '${role}'"
    sleep infinity & wait
}

# determine the action to perform
action="${1:-}"
case "${action}" in
    (--version)
        echo "${k8s_version_info[${k8s}]}"
        ;;

    (agent|server|dummy|"")
        simulate_service_running "${action:-no_role_specified}"
        ;;

    (*)
        echo "ERROR: Action '${action}' not implemented for simulated provider '${k8s}'"
        exit 1
        ;;
esac

# clean exit if we get here
exit 0
`
}

func setupSimulatedK8sEnv(t *testing.T) []string {
	err := os.MkdirAll(simulatedK8sDir, 0755)
	require.NoError(t, err, "Failed to create simulatedK8sDir %s directory", simulatedK8sDir)

	// write simulated provider script for currently simulated providers
	simulatedK8sScript := generateSimulatedK8sScript()
	k3sScriptPath := filepath.Join(simulatedK8sDir, "k3s")
	rke2ScriptPath := filepath.Join(simulatedK8sDir, "rke2")
	for _, k8sScript := range []string{k3sScriptPath, rke2ScriptPath} {
		err := os.WriteFile(k8sScript, []byte(simulatedK8sScript), 0755)
		require.NoError(t, err, "Failed to write simulated %s script", k8sScript)
	}

	services := map[string]string{
		// valid k3s services
		"k3s.service":        k3sScriptPath + " server",
		"k3s-agent.service":  k3sScriptPath + " agent",
		"k3s-custom.service": k3sScriptPath + " server",

		// valid rke2 services
		"rke2-server.service": rke2ScriptPath + " server",
		"rke2-agent.service":  rke2ScriptPath + " agent",

		// dummy services
		"k3s-dummy.service":  k3sScriptPath + " dummy",
		"rke2-dummy.service": rke2ScriptPath + " dummy",

		// invalid services
		"k3snodash.service":    k3sScriptPath,
		"k3s-invalid.service":  k3sScriptPath,
		"rke2-invalid.service": rke2ScriptPath,
	}

	etcSystemdDir := "/etc/systemd/system"
	var svcNames []string
	for name, execStart := range services {
		content := fmt.Sprintf("[Unit]\nDescription=Simulated %s\n\n[Service]\nExecStart=%s\n\n[Install]\nWantedBy=multi-user.target\n", name, execStart)
		err := os.WriteFile(filepath.Join(etcSystemdDir, name), []byte(content), 0644)
		require.NoError(t, err, "Failed to write simulated service %s", name)
		svcNames = append(svcNames, name)
	}

	systemdDaemonReload(t)

	return svcNames
}

func cleanupSimulatedK8sEnv(t *testing.T, services []string) {
	// skip test env cleanup if no cleanup env var exists
	if _, ok := os.LookupEnv(k8sProviderTestsNoCleanup); ok {
		return
	}

	for _, name := range services {
		// disable the service
		systemdServiceAction(t, name, "disable")

		// remove the simulated service
		os.Remove(filepath.Join(etcSystemdDir, name))
	}

	k3sScript := filepath.Join(simulatedK8sDir, "k3s")
	rke2Script := filepath.Join(simulatedK8sDir, "rke2")
	for _, k8sScript := range []string{k3sScript, rke2Script} {
		os.Remove(k8sScript)
	}

	systemdDaemonReload(t)
}

func systemdServiceAction(t *testing.T, name, action string) {
	// Enable/disable service immediately
	cli := helpers.NewRunner(t, "systemctl %s --now %s", action, name)
	cli.Run()
}

func systemdDaemonReload(t *testing.T) {
	// Enable/disable service immediately
	cli := helpers.NewRunner(t, "systemctl daemon-reload")
	cli.Run()
}

func TestKubernetesProviderInfo(t *testing.T) {
	assert := assert.New(t)

	// skip if systemctl not available
	if _, err := exec.LookPath("systemctl"); err != nil {
		t.Skip("Skipping feature test: systemctl not available in this environment")
	}

	// skip if provider tests enabled env var is not set
	if _, ok := os.LookupEnv(k8sProviderTestsEnabled); !ok {
		t.Skipf("Skipping feature test: env var %q not set", k8sProviderTestsEnabled)
	}

	services := setupSimulatedK8sEnv(t)
	defer cleanupSimulatedK8sEnv(t, services)

	targetServices := []struct {
		name     string
		role     string
		provider string
		version  string
		service  string
	}{
		{"RKE2 Server", "server", "rke2", rke2Version, "rke2-server.service"},
		{"RKE2 Agent", "agent", "rke2", rke2Version, "rke2-agent.service"},
		{"K3s Server", "server", "k3s", k3sVersion, "k3s.service"},
		{"K3s Agent", "agent", "k3s", k3sVersion, "k3s-agent.service"},
		{"K3s Custom Server", "server", "k3s", k3sVersion, "k3s-custom.service"},
	}
	for _, tt := range targetServices {
		t.Run(fmt.Sprintf("Detects %s", tt.name), func(t *testing.T) {
			systemdServiceAction(t, tt.service, "enable")
			defer systemdServiceAction(t, tt.service, "disable")

			expected := map[string]any{
				"type":    tt.provider,
				"role":    tt.role,
				"version": tt.version,
			}
			kp := getKubernetesProviderInfo(t)
			assert.NotNil(kp, "suseconnect --info should have returned a valid kubernetes_provider entry")
			assert.Equal(expected, kp, "suseconnect --info should have returned the expected kubernetes provider entry")
		})
	}

	t.Run("Ignores dummy roles", func(t *testing.T) {
		for _, svcName := range []string{"k3s-dummy.service", "rke2-dummy.service"} {
			systemdServiceAction(t, svcName, "enable")
			defer systemdServiceAction(t, svcName, "disable")
		}

		kp := getKubernetesProviderInfo(t)
		assert.Nil(kp, "Expected no valid kubernetes provider due to dummy roles")
	})

	t.Run("Ignores invalid execStart arguments", func(t *testing.T) {
		for _, svcName := range []string{"k3snodash.service", "k3s-invalid.service", "rke2-invalid.service"} {
			systemdServiceAction(t, svcName, "enable")
			defer systemdServiceAction(t, svcName, "disable")
		}

		kp := getKubernetesProviderInfo(t)
		assert.Nil(kp, "Expected no valid kubernetes provider due to invalid execStart arguments")
	})

	t.Run("Logs expected debug message when no providers detected", func(t *testing.T) {
		cli := helpers.NewRunner(t, "suseconnect --info --debug")
		cli.Run()
		assert.Equal(0, cli.ExitCode(), "suseconnect should not fail when multiple providers are enabled")
		assert.Contains(cli.Stderr(), "No kubernetes providers found", "suseconnect should log the expected debug message")
	})

	t.Run("Logs expected debug message when multiple providers detected", func(t *testing.T) {
		systemdServiceAction(t, "k3s.service", "enable")
		systemdServiceAction(t, "rke2-server.service", "enable")
		defer systemdServiceAction(t, "k3s.service", "disable")
		defer systemdServiceAction(t, "rke2-server.service", "disable")

		cli := helpers.NewRunner(t, "suseconnect --info --debug")
		cli.Run()
		assert.Equal(0, cli.ExitCode(), "suseconnect should not fail when multiple providers are enabled")
		assert.Contains(cli.Stderr(), "too many kubernetes providers enabled", "suseconnect should log the expected debug message")
	})

	t.Run("Successfully Skips Disabled Services", func(t *testing.T) {
		systemdServiceAction(t, "k3s-agent.service", "enable")
		kp := getKubernetesProviderInfo(t)
		assert.NotNil(kp)

		systemdServiceAction(t, "k3s-agent.service", "disable")
		kp = getKubernetesProviderInfo(t)
		assert.Nil(kp, "Disabled service was not correctly skipped")
	})
}
