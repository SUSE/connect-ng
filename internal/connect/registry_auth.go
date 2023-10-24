package connect

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// If the configuration file is not found for podman which resides
// in ${XDG_RUNTIME_DIR}/containers/auth.json it will fall back to
// docker configuration in ${HOME}/.docker.config.json
// See: https://docs.podman.io/en/stable/markdown/podman-login.1.html

const (
	DEFAULT_DOCKER_CLIENT_CONFIG = ".docker/config.json"
	DEFAULT_PODMAN_CONFIG        = "containers/auth.json"
	DEFAULT_SUSE_REGISTRY        = "https://registry.suse.com"
)

// we need this to allow mocking the system function in our tests
var (
	readFile  = os.ReadFile
	writeFile = os.WriteFile
	userHome  = os.UserHomeDir
	mkDirAll  = os.MkdirAll
)

// This is already implemented in containers. But for the fun sake we reinvent the wheel!
// See: https://github.com/containers/image/blob/main/pkg/docker/config/config.go
// Theoden: You have no dependency here!

type RegistryAuthConfig struct {
	AuthConfigs map[string]RegistryAuthentication `json:"auths"`
	CredHelpers map[string]string                 `json:"credHelpers,omitempty"`
}

type RegistryAuthentication struct {
	Auth          string `json:"auth,omitempty"`
	IdentityToken string `json:"identitytoken,omitempty"`
}

func newRegistryAuthConfig() *RegistryAuthConfig {
	return &RegistryAuthConfig{
		AuthConfigs: map[string]RegistryAuthentication{},
	}
}

func (cfg *RegistryAuthConfig) LoadFrom(path string) error {
	data, err := readFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, cfg)
}

func (cfg *RegistryAuthConfig) SaveTo(path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(path, data, 0600)
}

func (cfg *RegistryAuthConfig) isConfigured(registry string) bool {
	for configured := range cfg.AuthConfigs {
		if configured == registry {
			return true
		}
	}
	return false
}

func (cfg *RegistryAuthConfig) Set(registry string, login string, password string) {
	if cfg.isConfigured(registry) {
		Debug.Printf("`%s` is already configured. Skipping", registry)
		return
	}

	cred := fmt.Sprintf("%s:%s", login, password)
	auth := base64.StdEncoding.EncodeToString([]byte(cred))

	cfg.AuthConfigs[registry] = RegistryAuthentication{
		Auth: auth,
	}
}

func (cfg *RegistryAuthConfig) Get(registry string) (string, string, bool) {
	if cfg.isConfigured(registry) == false {
		return "", "", false
	}
	auth := cfg.AuthConfigs[registry].Auth

	decoded, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return "", "", false
	}

	result := strings.Split(string(decoded), ":")
	if len(result) != 2 {
		return "", "", false
	}
	return result[0], result[1], true
}

func (cfg *RegistryAuthConfig) Remove(registry string) {
	delete(cfg.AuthConfigs, registry)
}

func dockerConfigPath() (string, bool) {
	home, err := userHome()
	return filepath.Join(home, DEFAULT_DOCKER_CLIENT_CONFIG), err == nil
}

func podmanConfigPath() (string, bool) {
	path, found := os.LookupEnv("XDG_RUNTIME_DIR")
	return filepath.Join(path, DEFAULT_PODMAN_CONFIG), found
}

func setupRegistryAuthentication(login string, password string) {
	setup := func(pathFn func() (string, bool)) {
		config := newRegistryAuthConfig()

		if path, ok := pathFn(); ok {
			// This also fails if the file does not yet exist
			// so we continue to create it.
			if err := config.LoadFrom(path); err != nil {
				Debug.Printf("Could not load `%s`: %s", path, err)
			}
			config.Set(DEFAULT_SUSE_REGISTRY, login, password)

			if err := mkDirAll(filepath.Dir(path), 0777); err != nil {
				Debug.Printf("Could not create directory `%s`: %s", filepath.Dir(path), err)
				return
			}

			if err := config.SaveTo(path); err != nil {
				Debug.Printf("Could not save config to `%s`: %s", path, err)
			}
		}
	}
	setup(dockerConfigPath)
	setup(podmanConfigPath)
}

func removeRegistryAuthentication(login string, password string) {
	remove := func(pathFn func() (string, bool)) {
		config := newRegistryAuthConfig()

		if path, ok := pathFn(); ok {
			if err := config.LoadFrom(path); err != nil {
				Debug.Printf("Could not load `%s`: %s", path, err)
				return
			}
			l, p, found := config.Get(DEFAULT_SUSE_REGISTRY)

			// Make sure to only delete if the credentials actually match,
			// to not accidentially remove registry configuration which was
			// manually added
			if found == true && login == l && password == p {
				config.Remove(DEFAULT_SUSE_REGISTRY)

				if err := config.SaveTo(path); err != nil {
					Debug.Printf("Could not save config to `%s`: %s", path, err)
				}
			}
		}

	}
	remove(dockerConfigPath)
	remove(podmanConfigPath)
}
