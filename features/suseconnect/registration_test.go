package features

import (
	"testing"

	"github.com/SUSE/connect-ng/features/helpers"
	"github.com/stretchr/testify/assert"
)

func TestRegistration(t *testing.T) {
	t.Cleanup(helpers.CleanupPolutedFilesystem)
	t.Cleanup(helpers.TrySUSEConnectCleanup)
	t.Cleanup(helpers.TrySUSEConnectDeregister)

	t.Run("registering with invalid regcode", testRegisterWithInvalidRegcode)
	t.Run("registering with expired regcode", testRegisterWithExpiredRegcode)
	t.Run("registering with valid regcode", testRegisterWithValidRegcode)
	t.Run("deregistering", testDeregister)
	t.Run("registering and setting labels", testRegisterWithLabels)
}

func testRegisterWithInvalidRegcode(t *testing.T) {
	assert := assert.New(t)

	cli := helpers.NewRunner(t, "suseconnect -r SCRAMBLEMUMBLENOTEXISTENT")

	cli.Run()
	assert.Contains(cli.Stdout(), "Unknown Registration Code")
	assert.Equal(67, cli.ExitCode())
}

func testRegisterWithExpiredRegcode(t *testing.T) {
	assert := assert.New(t)

	env := helpers.NewEnv(t)
	cli := helpers.NewRunner(t, "suseconnect -r %s", env.EXPIRED_REGCODE)

	cli.Run()
	assert.Contains(cli.Stdout(), "Expired Registration Code.")
	assert.Equal(67, cli.ExitCode())
}

func testRegisterWithValidRegcode(t *testing.T) {
	assert := assert.New(t)

	env := helpers.NewEnv(t)
	cli := helpers.NewRunner(t, "suseconnect -r %s", env.REGCODE)

	cli.Run()
	assert.Contains(cli.Stdout(), "Registering system to SUSE Customer Center")
	assert.Contains(cli.Stdout(), "Successfully registered system")
	assert.FileExists(env.CredentialsPath("SCCcredentials"))
	assert.Equal(0, cli.ExitCode())
}

func testDeregister(t *testing.T) {
	assert := assert.New(t)
	env := helpers.NewEnv(t)
	cli := helpers.NewRunner(t, "suseconnect -d")

	cli.Run()
	assert.Contains(cli.Stdout(), "Successfully deregistered")
	assert.NoFileExists(env.CredentialsPath("SCCcredentials"))
	assert.Equal(0, cli.ExitCode())
}

func testRegisterWithLabels(t *testing.T) {
	assert := assert.New(t)

	env := helpers.NewEnv(t)
	api := helpers.NewValidationAPI(t)
	cli := helpers.NewRunner(t, "suseconnect -r %s --set-labels label-1,label-2", env.REGCODE)

	cli.Run()
	assert.Contains(cli.Stdout(), "Registering system to SUSE Customer Center")
	assert.Contains(cli.Stdout(), "Successfully registered system")

	api.FetchCredentials()

	assert.Equal(api.CurrentLabels(), []string{"label-1", "label-2"})
	assert.Equal(0, cli.ExitCode())
}
