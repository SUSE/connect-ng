package connect

const (
	defaultConfigPath = "/etc/SUSEConnect"
	defaultBaseURL    = "https://scc.suse.com/"
	defaultLang       = "en_US.UTF-8"
	defaultInsecure   = false
)

// Config holds the config!
type Config struct {
	Path     string
	BaseURL  string
	Language string
	Insecure bool
}

// LoadConfig reads the config from etc file
func LoadConfig() Config {
	// TODO actaully load the config from file

	// For PoC just use the defaults
	return Config{
		Path:     defaultConfigPath,
		BaseURL:  defaultBaseURL,
		Language: defaultLang,
		Insecure: defaultInsecure,
	}
}
