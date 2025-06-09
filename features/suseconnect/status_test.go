package features

import (
	"fmt"
	"testing"

	"github.com/SUSE/connect-ng/features/helpers"
	"github.com/stretchr/testify/assert"
)

func TestStatus(t *testing.T) {
	t.Cleanup(helpers.CleanupPolutedFilesystem)
	t.Cleanup(helpers.TrySUSEConnectCleanup)
	t.Cleanup(helpers.TrySUSEConnectDeregister)

	t.Run("JSON status when not registered", testStatusUnregistered)
	t.Run("Text status when not registered", testStatusTextUnregistered)
	t.Run("JSON status when registered", testStatusRegistered)
}

func testStatusUnregistered(t *testing.T) {
	assert := assert.New(t)

	zypp := helpers.NewZypper(t)
	cli := helpers.NewRunner(t, "suseconnect --status")
	cli.Run()

	identifier, version, arch := zypp.BaseProduct()
	expected := fmt.Sprintf(`[{"identifier":"%s","version":"%s","arch":"%s","status":"Not Registered"}]`, identifier, version, arch)

	assert.JSONEq(expected, cli.Stdout())
	assert.Equal(0, cli.ExitCode())
}

func testStatusTextUnregistered(t *testing.T) {
	assert := assert.New(t)

	zypp := helpers.NewZypper(t)
	cli := helpers.NewRunner(t, "suseconnect --status-text")
	cli.Run()

	identifier, version, arch := zypp.BaseProduct()
	triplet := fmt.Sprintf("%s/%s/%s", identifier, version, arch)

	assert.Contains(cli.Stdout(), "Installed Products:")
	assert.Contains(cli.Stdout(), triplet)
	assert.Contains(cli.Stdout(), "Not Registered")
	assert.Equal(0, cli.ExitCode())
}

func testStatusRegistered(t *testing.T) {
	assert := assert.New(t)

	env := helpers.NewEnv(t)
	api := helpers.NewValidationAPI(t)
	register := helpers.NewRunner(t, "suseconnect -r %s", env.REGCODE)

	register.Run()
	assert.Equal(0, register.ExitCode())
	api.FetchCredentials()
	activations := api.Activations()

	status := helpers.NewRunner(t, "suseconnect -s")
	status.Run()

	assert.Equal(0, status.ExitCode())
	json := helpers.AssertValidJSON[[]map[string]string](t, status.Stdout())
	enabled := []string{}

	for _, item := range json {
		if item["status"] == "Registered" {
			enabled = append(enabled, item["identifier"])
		}
	}

	for _, activation := range activations {
		assert.Contains(enabled, activation.Product.Identifier)
	}
}
