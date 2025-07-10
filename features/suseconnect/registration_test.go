package features

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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

func TestRegistrationWithProxies(t *testing.T) {
	t.Cleanup(helpers.CleanupPolutedFilesystem)
	t.Cleanup(helpers.TrySUSEConnectCleanup)
	t.Cleanup(helpers.TryCurlrcCleanup)

	t.Run("registering with HTTP_PROXY", testRegisterWithHttpProxy)
	t.Run("registering with HTTP_PROXY and $HOME/.curlrc", testRegisterProxyCurlrc)
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

func testRegisterWithHttpProxy(t *testing.T) {
	assert := assert.New(t)

	env := helpers.NewEnv(t)
	cli := helpers.NewRunner(t, "suseconnect -r %s", env.REGCODE)

	vars := os.Environ()
	vars = append(vars, "HTTPS_PROXY=https://invalid.proxy:1234")
	cli.Cmd.Env = vars

	cli.Run()
	assert.Contains(cli.Stdout(), "no such host")
	assert.Equal(1, cli.ExitCode())
}

func testRegisterProxyCurlrc(t *testing.T) {
	assert := assert.New(t)

	home := os.Getenv("HOME")
	assert.NotEmpty(home)

	// Generate the .curlrc file. It will be removed by a `Cleanup` function.
	err := os.WriteFile(filepath.Join(home, ".curlrc"), []byte("proxy-user = \"user:password\""), 0644)
	assert.NoError(err)

	// An HTTP proxy will be listening to the request sent by `cli.Run()`.
	handler := &helpers.Proxy{Assert: assert, ExpectedProxyAuth: "user:password"}
	go func() {
		if err := http.ListenAndServe("0.0.0.0:8080", handler); err != nil {
			assert.Fail(fmt.Sprintf("could not create proxy server: %v", err))
		}
	}()

	env := helpers.NewEnv(t)
	cli := helpers.NewRunner(t, "suseconnect -r %s", env.REGCODE)

	vars := os.Environ()
	vars = append(vars, "HTTPS_PROXY=0.0.0.0:8080")
	cli.Cmd.Env = vars
	cli.Run()

	// NOTE: still fails because this proxy is returning an HTTP response for an
	// HTTPS server.
	assert.Contains(cli.Stdout(), "server gave HTTP response to HTTPS client")
	assert.Equal(1, cli.ExitCode())
}
