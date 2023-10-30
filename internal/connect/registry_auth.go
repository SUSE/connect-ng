package connect

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// If the configuration file is not found for podman which resides
// in ${XDG_RUNTIME_DIR}/containers/auth.json it will fall back to
// docker configuration in ${HOME}/.docker/config.json
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
	stat      = syscall.Stat
	chown     = syscall.Chown
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

func dockerConfigPath() (string, int, int, bool) {
	home, err := userHome()
	uId, gId := getPathOwnership(home)

	return filepath.Join(home, DEFAULT_DOCKER_CLIENT_CONFIG), uId, gId, err == nil
}

func podmanConfigPath() (string, int, int, bool) {
	// In this case usually XDG_RUNTIME_DIR is set to a
	// login user and NOT to the calling user (root)
	// This way we need to fetch the user id and group id
	// by looking into the ownership of the runtime path
	path, found := os.LookupEnv("XDG_RUNTIME_DIR")
	uId, gId := getPathOwnership(path)

	return filepath.Join(path, DEFAULT_PODMAN_CONFIG), uId, gId, found
}

func getPathOwnership(path string) (int, int) {
	fileStat := syscall.Stat_t{}

	if err := stat(path, &fileStat); err != nil {
		// we assume the user root
		return 0, 0
	}
	return int(fileStat.Uid), int(fileStat.Gid)
}

func setupRegistryAuthentication(login string, password string) {
	setup := func(pathFn func() (string, int, int, bool)) {
		config := newRegistryAuthConfig()

		if path, uId, gId, ok := pathFn(); ok {
			dir := filepath.Dir(path)
			Info.Printf("dir: %s", dir)

			// This also fails if the file does not yet exist
			// so we continue to create it.
			if err := config.LoadFrom(path); err != nil {
				Info.Printf("Could not load `%s`: %s", path, err)
			}
			config.Set(DEFAULT_SUSE_REGISTRY, login, password)

			if err := mkDirAll(dir, 0775); err != nil {
				Info.Printf("Could not create directory `%s`: %s", dir, err)
				return
			}

			if err := config.SaveTo(path); err != nil {
				Info.Printf("Could not save config to `%s`: %s", path, err)
				return
			}
			chown(dir, uId, gId)
			chown(path, uId, gId)
		}
	}
	setup(dockerConfigPath)
	setup(podmanConfigPath)
}

func removeRegistryAuthentication(login string, password string) {
	remove := func(pathFn func() (string, int, int, bool)) {
		config := newRegistryAuthConfig()

		if path, _, _, ok := pathFn(); ok {
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
