package features

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SUSE/connect-ng/features/helpers"
	"github.com/stretchr/testify/assert"
)

func TestChangedRoot(t *testing.T) {
	assert := assert.New(t)
	env := helpers.NewEnv(t)

	root := helpers.SetupCustomRoot(t)
	t.Cleanup(helpers.CleanupAll)
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
