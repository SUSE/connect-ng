package features

import (
	"testing"

	"github.com/SUSE/connect-ng/features/helpers"
	"github.com/stretchr/testify/assert"
)

func TestRollback(t *testing.T) {
	t.Cleanup(helpers.CleanupPolutedFilesystem)
	t.Cleanup(helpers.TrySUSEConnectCleanup)
	t.Cleanup(helpers.TrySUSEConnectDeregister)

	t.Run("rollback and reset state", testRollbackToNormal)
}

func testRollbackToNormal(t *testing.T) {
	assert := assert.New(t)

	env := helpers.NewEnv(t)
	api := helpers.NewValidationAPI(t)
	register := helpers.NewRunner(t, "suseconnect -r %s", env.REGCODE)

	register.Run()
	assert.Equal(0, register.ExitCode())
	api.FetchCredentials()

	activations := api.Activations()
	for i, activation := range activations {
		// Remove every other product credentials file
		if i%2 > 0 {
			path := env.CredentialsPath(helpers.FriendlyNameToCredentialsName(activation.Product.FriendlyName))
			helpers.RemoveFile(t, path)
			assert.NoFileExists(path)
		}
	}

	rollback := helpers.NewRunner(t, "suseconnect --rollback")
	rollback.Run()

	assert.Contains(rollback.Stdout(), "Starting to sync system product activations to the server. This can take some time...")
	assert.Equal(0, rollback.ExitCode())

	for _, activation := range activations {
		path := env.CredentialsPath(helpers.FriendlyNameToCredentialsName(activation.Product.FriendlyName))
		assert.FileExists(path)
	}
}
