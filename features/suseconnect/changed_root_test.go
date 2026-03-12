package features

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/SUSE/connect-ng/features/helpers"
	"github.com/stretchr/testify/assert"
)

func TestChangedRoot(t *testing.T) {
	assert := assert.New(t)
	env := helpers.NewEnv(t)

	root := setupCustomRoot(t)
	namespace := "test-namespace"

	// No need to install rpm packages just registration
	t.Setenv("SKIP_SERVICE_INSTALL", "true")

	cli := helpers.NewRunner(t, "suseconnect --root %s --namespace %s -r %s", root, namespace, env.REGCODE)
	cli.Run()

	assert.Equal(0, cli.ExitCode())
	assert.Contains(cli.Stdout(), "Rooted at: "+root)

	t.Run("check configuration location", func(t *testing.T) {
		testChangedRootConfigLocation(t, root, namespace)
	})

	t.Run("check credentials location", func(t *testing.T) {
		testChangedRootCredentialsLocation(t, root)
	})
}

func setupCustomRoot(t *testing.T) string {
	assert := assert.New(t)
	root := t.TempDir()

	err := os.MkdirAll(filepath.Join(root, "etc"), 0755)
	assert.NoError(err)

	// FIXME: Once golang 1.23 is integrated this becomes much more nice to implement in pure go
	err = exec.Command("cp", "-r", "/etc/zypp", filepath.Join(root, "etc/zypp")).Run()
	assert.NoError(err)

	err = exec.Command("cp", "-r", "/etc/products.d", filepath.Join(root, "etc/products.d")).Run()
	assert.NoError(err)

	return root
}

func testChangedRootConfigLocation(t *testing.T, root string, namespace string) {
	assert := assert.New(t)

	config := filepath.Join(root, "/etc/SUSEConnect")
	assert.FileExists(config)

	content, err := os.ReadFile(config)
	assert.NoError(err)
	assert.Contains(string(content), "namespace: "+namespace)
}

func testChangedRootCredentialsLocation(t *testing.T, root string) {
	assert := assert.New(t)

	credentials := filepath.Join(root, "/etc/zypp/credentials.d/SCCcredentials")
	assert.FileExists(credentials)
}
