package features

import (
	"regexp"
	"testing"

	"github.com/SUSE/connect-ng/features/helpers"
	"github.com/stretchr/testify/assert"
)

func TestCLI(t *testing.T) {
	t.Run("showing help", testCLIPrintsHelp)
	t.Run("showing version", testCLIPrintsVersion)
}

func testCLIPrintsHelp(t *testing.T) {
	assert := assert.New(t)
	cli := helpers.NewRunner(t, "suseconnect --help")

	cli.Run()
	assert.Contains(cli.Stdout(), "Register SUSE Linux Enterprise installations with the SUSE Customer Center")
	assert.Contains(cli.Stdout(), "-r, --regcode [REGCODE]")
	assert.Contains(cli.Stdout(), "-d, --de-register")
	assert.Contains(cli.Stdout(), "-p, --product [PRODUCT]")
	assert.Contains(cli.Stdout(), "-s, --status")
	assert.Contains(cli.Stdout(), "-l, --list-extensions")
	assert.Contains(cli.Stdout(), "-i, --info")
	assert.Contains(cli.Stdout(), "--url [URL]")
	assert.Contains(cli.Stdout(), "--set-labels [LABELS]")

	assert.Contains(cli.Stdout(), "--debug")
	assert.Contains(cli.Stdout(), "--json")
	assert.Contains(cli.Stdout(), "--gpg-auto-import-keys")
	assert.Equal(0, cli.ExitCode())

}

func testCLIPrintsVersion(t *testing.T) {
	assert := assert.New(t)
	cli := helpers.NewRunner(t, "suseconnect --version")

	cli.Run()
	assert.Regexp(regexp.MustCompile(`\d{1,2}\.\d{1,2}(\.\d{1,2})?`), cli.Stdout())
	assert.Equal(0, cli.ExitCode())
}
