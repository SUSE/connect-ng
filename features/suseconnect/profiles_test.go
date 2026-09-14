package features

import (
	"testing"

	"github.com/SUSE/connect-ng/features/helpers"
	"github.com/stretchr/testify/assert"
)

const configFileText = `url: https://scc.suse.com
language: en
insecure: false
auto_agree_with_licenses: false
enable_system_uptime_tracking: false
no_zypper_refs: false
collectors:
  rpm_packages:
    state: enabled
`

func TestProfiles(t *testing.T) {
	t.Cleanup(helpers.CleanupPolutedFilesystem)
	t.Cleanup(helpers.TrySUSEConnectCleanup)
	t.Cleanup(helpers.TrySUSEConnectDeregister)

	t.Run("Test Profile Creation", testProfilesCreate)
	t.Run("Test Profile Cleanup", testProfilesCleanup)
}

func testProfilesCreate(t *testing.T) {
	assert := assert.New(t)

	env := helpers.NewEnv(t)
	helpers.CreateFileWithContent("/etc/SUSEConnect", configFileText) 
	cli := helpers.NewRunner(t, "suseconnect -r %s", env.REGCODE)

	cli.Run()
	assert.Equal(0, cli.ExitCode())
	assert.FileExists("/run/suseconnect/kernel-modules-profile-id")
	assert.FileExists("/run/suseconnect/pci-data-profile-id")
	assert.FileExists("/run/suseconnect/pkgs.txt")
	assert.FileExists("/run/suseconnect/clear-cache-count")
}

func testProfilesCleanup(t *testing.T) {
	assert := assert.New(t)

	cli := helpers.NewRunner(t, "suseconnect -d")

	cli.Run()
	assert.Equal(0, cli.ExitCode())
	assert.NoFileExists("/run/suseconnect/kernel-modules-profile-id")
	assert.NoFileExists("/run/suseconnect/pci-data-profile-id")
	assert.NoFileExists("/run/suseconnect/pkgs.txt")
	assert.FileExists("/run/suseconnect/clear-cache-count")
	helpers.DeleteFile("/etc/SUSEConnect")
}

