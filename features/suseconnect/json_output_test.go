package features

import (
	"testing"

	"github.com/SUSE/connect-ng/features/helpers"
	"github.com/stretchr/testify/assert"
)

func TestJSONOutput(t *testing.T) {
	t.Cleanup(helpers.CleanupPolutedFilesystem)
	t.Cleanup(helpers.TrySUSEConnectCleanup)
	t.Cleanup(helpers.TrySUSEConnectDeregister)

	t.Run("register with output set to JSON without REGCODE", testRegisterJSONInvalidRegcode)
	t.Run("registering with output set to JSON", testRegisterJSON)
	t.Run("deregister with output set to JSON", testDeregisterJSON)
	t.Run("deregister with output set to JSON but not registered", testDeregisterUnregisteredJSON)
}

func testRegisterJSONInvalidRegcode(t *testing.T) {
	assert := assert.New(t)

	cli := helpers.NewRunner(t, "suseconnect --json -r SCRUMBLEMUMBLENOTEXISTENT")
	cli.Run()

	json := helpers.AssertValidJSON[map[string]any](t, cli.Stdout())

	assert.False(json["success"].(bool))
	assert.Contains(json["message"], "Error: Registration server returned 'Unknown Registration Code.'")

	// FIXME: This is actually wrong and should be 67! Fix the implementation!
	assert.Equal(1, cli.ExitCode())
}

func testRegisterJSON(t *testing.T) {
	assert := assert.New(t)

	env := helpers.NewEnv(t)
	cli := helpers.NewRunner(t, "suseconnect --json -r %s", env.REGCODE)
	cli.Run()

	json := helpers.AssertValidJSON[map[string]any](t, cli.Stdout())

	assert.True(json["success"].(bool))
	assert.Contains(json["message"], "Successfully registered system")
	assert.Equal(0, cli.ExitCode())
}

func testDeregisterJSON(t *testing.T) {
	assert := assert.New(t)

	cli := helpers.NewRunner(t, "suseconnect --json -d")
	cli.Run()

	json := helpers.AssertValidJSON[map[string]any](t, cli.Stdout())

	assert.True(json["success"].(bool))
	assert.Contains(json["message"], "Successfully deregistered system")
	assert.Equal(0, cli.ExitCode())
}

func testDeregisterUnregisteredJSON(t *testing.T) {
	assert := assert.New(t)

	cli := helpers.NewRunner(t, "suseconnect --json -d")
	cli.Run()

	json := helpers.AssertValidJSON[map[string]any](t, cli.Stdout())

	assert.False(json["success"].(bool))
	assert.Contains(json["message"], "System not registered")

	// FIXME: This is actually wrong and should be 69. Fix the implementation
	assert.Equal(1, cli.ExitCode())
}
