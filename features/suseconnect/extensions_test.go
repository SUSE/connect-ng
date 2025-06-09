package features

import (
	"fmt"
	"testing"

	"github.com/SUSE/connect-ng/features/helpers"
	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/stretchr/testify/assert"
)

func TestExtensions(t *testing.T) {
	t.Cleanup(helpers.CleanupPolutedFilesystem)
	t.Cleanup(helpers.TrySUSEConnectCleanup)
	t.Cleanup(helpers.TrySUSEConnectDeregister)

	t.Run("list extensions without registration", testListExtensionsWithoutRegistration)
	t.Run("register and activate recommended modules", testRegisterWithExtensions)
	t.Run("list extensions", testListExtensions)
	t.Run("activate and deactivate additional module", testActivateAndDeactivateModule)
	t.Run("activate extension without regcode", testActivateExtensionWithoutRegcode)
	t.Run("activate and deactivate extension with regcode", testActivateAndDeactivateExtensionWithRegcode)
	t.Run("activate leaf module", testActivateLeafModule)
}

func testListExtensionsWithoutRegistration(t *testing.T) {
	assert := assert.New(t)
	cli := helpers.NewRunner(t, "suseconnect --list-extensions")
	cli.Run()

	assert.Contains(cli.Stdout(), "To list extensions, you must first register the base product")
	assert.Equal(1, cli.ExitCode())
}

func testRegisterWithExtensions(t *testing.T) {
	assert := assert.New(t)

	env := helpers.NewEnv(t)
	cli := helpers.NewRunner(t, "suseconnect -r %s", env.REGCODE)
	api := helpers.NewValidationAPI(t)
	zypper := helpers.NewZypper(t)

	id, version, arch := zypper.BaseProduct()
	cli.Run()

	assert.Contains(cli.Stdout(), fmt.Sprintf("Activating %s %s %s ...", id, version, arch))

	api.FetchCredentials()
	// Fetch all possible free and recommended extensions from the API
	tree := api.ProductTree(id, version, arch)

	// Make sure all recommended extensions are activated
	tree.TraverseExtensions(func(ext registration.Product) (bool, error) {
		lookForMsg := fmt.Sprintf("recommended %s is activated", ext.Identifier)
		needle := fmt.Sprintf("Activating %s %s %s ...", ext.Identifier, ext.Version, ext.Arch)

		if ext.Free && ext.Recommended {
			assert.Contains(cli.Stdout(), needle, lookForMsg)
		}
		return true, nil
	})

	assert.Equal(0, cli.ExitCode())
}

func testListExtensions(t *testing.T) {
	assert := assert.New(t)

	cli := helpers.NewRunner(t, "suseconnect --list-extensions")
	api := helpers.NewValidationAPI(t)
	zypper := helpers.NewZypper(t)

	cli.Run()
	api.FetchCredentials()

	tree := api.ProductTree(zypper.BaseProduct())

	tree.TraverseExtensions(func(ext registration.Product) (bool, error) {
		assert.Contains(cli.Stdout(), ext.FriendlyName)

		needle := fmt.Sprintf("-p %s/%s/%s", ext.Identifier, ext.Version, ext.Arch)
		if ext.Free {
			assert.Contains(cli.Stdout(), needle)
		} else {
			// Note: Stripping terminal escape codes is no trivial task...
			assert.Contains(cli.Stdout(), fmt.Sprintf("%s -r \x1b[32m\x1b[1mADDITIONAL REGCODE\x1b[0m", needle))
		}
		return true, nil
	})
}

func testActivateAndDeactivateModule(t *testing.T) {
	assert := assert.New(t)

	zypp := helpers.NewZypper(t)
	api := helpers.NewValidationAPI(t)
	// Assume we are already registered
	api.FetchCredentials()

	tree := api.ProductTree(zypp.BaseProduct())
	tree.TraverseExtensions(func(ext registration.Product) (bool, error) {
		// Try to activate all easily reachable modules and activate and deactivate them
		if ext.Free && !ext.Recommended {
			fmt.Printf("[info] Try activating %s\n", ext.FriendlyName)
			activate := helpers.NewRunner(t, "suseconnect -p %s/%s/%s", ext.Identifier, ext.Version, ext.Arch)
			deactivate := helpers.NewRunner(t, "suseconnect -d -p %s/%s/%s", ext.Identifier, ext.Version, ext.Arch)

			activate.Run()
			assert.Contains(activate.Stdout(), fmt.Sprintf("Activating %s %s %s ...", ext.Identifier, ext.Version, ext.Arch))
			assert.True(zypp.ServiceExists(ext))
			assert.Equal(0, activate.ExitCode())

			deactivate.Run()
			assert.Contains(deactivate.Stdout(), fmt.Sprintf("Deactivating %s %s %s ...", ext.Identifier, ext.Version, ext.Arch))
			assert.False(zypp.ServiceExists(ext))
			assert.Equal(0, deactivate.ExitCode())
		}
		return false, nil
	})
}

func testActivateExtensionWithoutRegcode(t *testing.T) {
	assert := assert.New(t)

	zypp := helpers.NewZypper(t)
	_, version, arch := zypp.BaseProduct()

	// Pick an extension which requires and additional registration code to be activated
	extension := fmt.Sprintf("%s/%s/%s", "sle-ha", version, arch)
	cli := helpers.NewRunner(t, "suseconnect -p %s", extension)

	cli.Run()
	assert.Contains(cli.Stdout(), "Error: Registration server returned 'Please provide a Registration Code for this product'")
	assert.Equal(67, cli.ExitCode())
}

func testActivateAndDeactivateExtensionWithRegcode(t *testing.T) {
	assert := assert.New(t)

	env := helpers.NewEnv(t)
	zypp := helpers.NewZypper(t)
	_, version, arch := zypp.BaseProduct()

	// Pick an extension which requires and additional registration code to be activated
	extension := fmt.Sprintf("%s/%s/%s", "sle-ha", version, arch)

	activate := helpers.NewRunner(t, "suseconnect -p %s -r %s", extension, env.HA_REGCODE)
	deactivate := helpers.NewRunner(t, "suseconnect -d -p %s", extension)

	activate.Run()
	assert.Contains(activate.Stdout(), fmt.Sprintf("Activating sle-ha %s %s ...", version, arch))
	assert.Equal(0, activate.ExitCode())

	deactivate.Run()
	assert.Contains(deactivate.Stdout(), fmt.Sprintf("Deactivating sle-ha %s %s ...", version, arch))
	assert.Equal(0, deactivate.ExitCode())
}

func testActivateLeafModule(t *testing.T) {
	assert := assert.New(t)

	zypp := helpers.NewZypper(t)
	_, version, arch := zypp.BaseProduct()

	// Pick a module which is depended on another module which is not yet activated
	// HPC requires Web and Scripting module
	extension := fmt.Sprintf("%s/%s/%s", "sle-module-hpc", version, arch)

	cli := helpers.NewRunner(t, "suseconnect -p %s", extension)
	cli.Run()

	assert.Contains(cli.Stdout(), "Error: Registration server returned 'The product you are attempting to activate (HPC Module 15 SP6 x86_64) requires one of these products to be activated first: Web and Scripting Module 15 SP6 x86_64'")
	assert.Equal(67, cli.ExitCode())
}
