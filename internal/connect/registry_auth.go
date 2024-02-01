package connect

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NOTE: Podman will read the same configuration file as the docker cli.
//       We utilize this to not have to write multiple files. Since the default
//       config is set to $XDG_RUNTIME_DIR/container/auth.json which should considered
//       as volatile storage.

const (
	DEFAULT_DOCKER_CLIENT_CONFIG = ".docker/config.json"
	DEFAULT_SUSE_REGISTRY        = "https://registry.suse.com"
)

// we need this to allow mocking the system function in our tests
var (
	readFile  = os.ReadFile
	writeFile = os.WriteFile
	userHome  = os.UserHomeDir
	mkDirAll  = os.MkdirAll
)

// NOTE: We only need a fraction of potential data supplied by the docker configuration file.
//       But since we need to read and later write the file we must keep the existing data
//       and write it to the new file, to not lose information.
//
//       There is an alternative implementation idea of using a tree like structure but
//       for the sake of simplicity the configuration structure is replicated.
//
//       See: https://github.com/docker/cli/blob/master/cli/config/configfile/file.go#L18

type RegistryAuthConfig struct {
	AuthConfigs          map[string]RegistryAuthentication `json:"auths"`
	HTTPHeaders          map[string]string                 `json:"HttpHeaders,omitempty"`
	PsFormat             string                            `json:"psFormat,omitempty"`
	ImagesFormat         string                            `json:"imagesFormat,omitempty"`
	NetworksFormat       string                            `json:"networksFormat,omitempty"`
	PluginsFormat        string                            `json:"pluginsFormat,omitempty"`
	VolumesFormat        string                            `json:"volumesFormat,omitempty"`
	StatsFormat          string                            `json:"statsFormat,omitempty"`
	DetachKeys           string                            `json:"detachKeys,omitempty"`
	CredentialsStore     string                            `json:"credsStore,omitempty"`
	CredentialHelpers    map[string]string                 `json:"credHelpers,omitempty"`
	Filename             string                            `json:"-"` // Note: for internal use only
	ServiceInspectFormat string                            `json:"serviceInspectFormat,omitempty"`
	ServicesFormat       string                            `json:"servicesFormat,omitempty"`
	TasksFormat          string                            `json:"tasksFormat,omitempty"`
	SecretFormat         string                            `json:"secretFormat,omitempty"`
	ConfigFormat         string                            `json:"configFormat,omitempty"`
	NodesFormat          string                            `json:"nodesFormat,omitempty"`
	PruneFilters         []string                          `json:"pruneFilters,omitempty"`
	Proxies              map[string]string                 `json:"proxies,omitempty"`
	Experimental         string                            `json:"experimental,omitempty"`
	CurrentContext       string                            `json:"currentContext,omitempty"`
	CLIPluginsExtraDirs  []string                          `json:"cliPluginsExtraDirs,omitempty"`
	Plugins              map[string]map[string]string      `json:"plugins,omitempty"`
	Aliases              map[string]string                 `json:"aliases,omitempty"`
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

func setupRegistryAuthentication(login string, password string) {
	config := newRegistryAuthConfig()

	if base, err := userHome(); err == nil {
		path := filepath.Join(base, DEFAULT_DOCKER_CLIENT_CONFIG)
		dir := filepath.Dir(path)

		// This also fails if the file does not yet exist
		// so we continue to create it.
		if err := config.LoadFrom(path); err != nil {
			Debug.Printf("Could not load `%s`: %s", path, err)
		}
		config.Set(DEFAULT_SUSE_REGISTRY, login, password)

		if err := mkDirAll(dir, 0775); err != nil {
			Debug.Printf("Could not create directory `%s`: %s", dir, err)
			return
		}

		if err := config.SaveTo(path); err != nil {
			Debug.Printf("Could not save config to `%s`: %s", path, err)
			return
		}

		Debug.Printf("SUSE registry system authentication written to `%s`", path)
	}
}

func removeRegistryAuthentication(login string, password string) {
	config := newRegistryAuthConfig()

	if home, err := userHome(); err == nil {
		path := filepath.Join(home, DEFAULT_DOCKER_CLIENT_CONFIG)

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

			Debug.Printf("SUSE registry system authentication removed from `%s`", path)
		}
	}

}
