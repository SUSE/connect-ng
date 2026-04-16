package features

import (
	"path/filepath"
	"testing"

	"github.com/SUSE/connect-ng/features/helpers"
	"github.com/stretchr/testify/assert"
)

func TestInstallPackage(t *testing.T) {
	t.Run("install PackageHub with auto-import keys - default root", testInstallPackageHubDefaultRoot)
	t.Run("install PackageHub with auto-import keys - custom root", testInstallPackageHubCustomRoot)
	t.Run("skip installation if package already installed - custom root", testSkipInstallationIfInstalledCustomRoot)
	t.Run("install release package with default root", testInstallReleasePackageDefaultRoot)
	t.Run("install release package with custom root", testInstallReleasePackageCustomRoot)
}

func testInstallPackageHubDefaultRoot(t *testing.T) {
	t.Cleanup(helpers.CleanupAll)

	assert := assert.New(t)
	env := helpers.NewEnv(t)

	// Register the system
	register := helpers.NewRunner(t, "suseconnect -r %s", env.REGCODE)
	register.Run()
	assert.Equal(0, register.ExitCode())

	zypp := helpers.NewZypper(t)
	_, version, arch := zypp.BaseProduct()

	extension := "PackageHub"
	activate := helpers.NewRunner(t, "suseconnect -p %s/%s/%s --gpg-auto-import-keys", extension, version, arch)
	activate.Run()

	if activate.ExitCode() == 0 {
		assert.Contains(activate.Stdout(), "Activating")
	} else {
		// If it fails, it should not be due to signature verification
		assert.NotContains(activate.Stderr(), "signature verification failed")
	}
}

func testInstallPackageHubCustomRoot(t *testing.T) {
	assert := assert.New(t)
	env := helpers.NewEnv(t)

	root := helpers.SetupCustomRoot(t)
	t.Cleanup(helpers.CleanupAll)
	namespace := "test-packagehub"

	// Register with custom root and auto-import keys
	register := helpers.NewRunner(t, "suseconnect --root %s --namespace %s -r %s --gpg-auto-import-keys", root, namespace, env.REGCODE)
	register.Run()

	if register.ExitCode() == 0 {
		assert.Contains(register.Stdout(), "Rooted at: "+root)
		assert.Contains(register.Stdout(), "Activating")
	} else {
		// If it fails, it should not be due to signature verification
		assert.NotContains(register.Stderr(), "signature verification failed")
	}

	// Verify the root flag was properly used
	credPath := filepath.Join(root, "etc/zypp/credentials.d/SCCcredentials")
	assert.FileExists(credPath, "Credentials should be in custom root")
}

func testSkipInstallationIfInstalledCustomRoot(t *testing.T) {
	assert := assert.New(t)
	env := helpers.NewEnv(t)

	root := helpers.SetupCustomRoot(t)
	t.Cleanup(helpers.CleanupAll)
	namespace := "test-skip-install"

	// First registration
	register1 := helpers.NewRunner(t, "suseconnect --root %s --namespace %s -r %s --gpg-auto-import-keys", root, namespace, env.REGCODE)
	register1.Run()
	assert.Contains(register1.Stdout(), "Installing release package")
	assert.Equal(0, register1.ExitCode())

	// Second registration with same root - should not fail
	// NOTE: registering without deregistering will orphan the previous registration in
	// SCC, leaving us unable to clean it up.
	deregister1 := helpers.NewRunner(t, "suseconnect --root %s --namespace %s -d", root, namespace)
	deregister1.Run()

	register2 := helpers.NewRunner(t, "suseconnect --root %s --namespace %s -r %s", root, namespace, env.REGCODE)
	register2.Run()
	assert.Contains(register2.Stdout(), "Installing release package")
	assert.Equal(0, register2.ExitCode(), "Re-registration with custom root should succeed")

	// Config should exist
	configPath := filepath.Join(root, "/etc/SUSEConnect")
	assert.FileExists(configPath)
}

func testInstallReleasePackageDefaultRoot(t *testing.T) {
	t.Cleanup(helpers.CleanupAll)

	assert := assert.New(t)
	env := helpers.NewEnv(t)

	zypp := helpers.NewZypper(t)
	namespace := "test-install-release-package"

	extension := "sle-module-basesystem"
	cli := helpers.NewRunner(t, "suseconnect --namespace %s -r %s", namespace, env.REGCODE)
	cli.Run()

	assert.Equal(0, cli.ExitCode())
	assert.Contains(cli.Stdout(), "Successfully registered system")

	// Verify the release package was installed by checking for the product
	products := zypp.FetchProducts()
	foundProduct := false
	for _, product := range products {
		if product.Identifier == extension {
			foundProduct = true
			break
		}
	}
	assert.True(foundProduct, "Release package should have installed the product")
}

func testInstallReleasePackageCustomRoot(t *testing.T) {
	assert := assert.New(t)
	env := helpers.NewEnv(t)

	// Setup custom root
	root := helpers.SetupCustomRoot(t)
	t.Cleanup(helpers.CleanupAll)
	zypp := helpers.NewZypper(t)
	namespace := "test-install-release-package"

	extension := "sle-module-basesystem"

	// Register with custom root
	register := helpers.NewRunner(t, "suseconnect --root %s --namespace %s -r %s --gpg-auto-import-keys", root, namespace, env.REGCODE)
	register.Run()

	assert.Equal(0, register.ExitCode(), "Registration with custom root should succeed")
	assert.Contains(register.Stdout(), "Rooted at: "+root)
	assert.Contains(register.Stdout(), "Successfully registered system")

	// Verify the release package was installed by checking for the product
	products := zypp.FetchProducts()
	foundProduct := false
	for _, product := range products {
		if product.Identifier == extension {
			foundProduct = true
			break
		}
	}
	assert.True(foundProduct, "Release package should have installed the product with custom root")

	configPath := filepath.Join(root, "/etc/SUSEConnect")
	assert.FileExists(configPath, "Config should exist in custom root")
}
