package features

import (
	"testing"

	"github.com/SUSE/connect-ng/features/helpers"
	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	t.Run("shows collected information as JSON blob", testInfoPrintsJSON)
}

func testInfoPrintsJSON(t *testing.T) {
	assert := assert.New(t)

	cli := helpers.NewRunner(t, "suseconnect --info")

	cli.Run()
	json := helpers.AssertValidJSON[map[string]any](t, cli.Stdout())

	assert.Contains(json, "cpus")
	assert.Contains(json, "mem_total")
	assert.Contains(json, "sockets")
	assert.Equal(0, cli.ExitCode())
}
