package connect

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/SUSE/connect-ng/internal/collectors"
	"github.com/SUSE/connect-ng/internal/util"
	collectorsconfig "github.com/SUSE/connect-ng/pkg/collectors"
	"github.com/SUSE/connect-ng/pkg/registration"
	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigPath                 = "/etc/SUSEConnect"
	defaultBaseURL                    = "https://scc.suse.com"
	defaultInsecure                   = false
	defaultSkip                       = false
	defaultEnableSystemUptimeTracking = false
)

// Kinds of servers which are supported by SUSEConnect.
type ServerType uint64

const (
	UnknownProvider ServerType = iota
	SccProvider
	RmtProvider
)

// OutputKind is an enum that describes which kind of output is expected from a
// CLI point of view.
type OutputKind int

const (
	// All output from the CLI is to be given in clear text.
	Text OutputKind = iota

	// All output from the CLI is to be given as JSON blobs.
	JSON
)

type Options struct {
	Path                       string
	BaseURL                    string `json:"url" yaml:"url"`
	Language                   string `json:"language" yaml:"language"`
	Insecure                   bool   `json:"insecure" yaml:"insecure"`
	Namespace                  string `json:"namespace" yaml:"namespace"`
	FsRoot                     string
	Token                      string
	Product                    registration.Product
	InstanceDataFile           string
	Email                      string `json:"email" yaml:"email"`
	AutoAgreeEULA              bool   `yaml:"auto_agree_with_licenses"`
	EnableSystemUptimeTracking bool   `yaml:"enable_system_uptime_tracking"`
	ServerType                 ServerType
	NoZypperRefresh            bool `yaml:"no_zypper_refs"`
	AutoImportRepoKeys         bool
	SkipServiceInstall         bool
	OutputKind                 OutputKind
}

// configFile represents the structure of the YAML configuration file
type configFile struct {
	Options    `yaml:",inline"`                // Flattened top-level options
	Collectors map[string]collectorConfigEntry `yaml:"collectors"`
}

// collectorConfigEntry represents a single collector's configuration
type collectorConfigEntry struct {
	Enabled *bool `yaml:"enabled"` // Pointer distinguishes unset vs. false
}

// Returns the Options suitable for targeting the SCC reference server without a
// proxy.
func DefaultOptions() *Options {
	return &Options{
		Path:                       DefaultConfigPath,
		BaseURL:                    defaultBaseURL,
		Insecure:                   defaultInsecure,
		SkipServiceInstall:         defaultSkip,
		EnableSystemUptimeTracking: defaultEnableSystemUptimeTracking,
		ServerType:                 UnknownProvider,
	}
}

// Returns the name of the server as expected by CLI tools.
func (opts *Options) ServerName() string {
	if opts.IsScc() {
		return "SUSE Customer Center"
	}
	return "registration proxy " + opts.BaseURL
}

// Prints the given message on `Info` or `Debug` depending on the OutputKind.
func (opts *Options) Print(msg string) {
	switch opts.OutputKind {
	case Text:
		util.Info.Printf(msg)
	case JSON:
		util.Debug.Printf(msg)
	}
}

// Returns an Options object which has the default values from `DefaultOptions`
// but fills the relevant data from the configuration as given by its `path`.
//
// If this path does not exist, then it takes the defaults. If this path does
// exist but there are problems when parsing it, then an error will be returned.
func ReadFromConfiguration(path string) (*Options, error) {
	cfg := DefaultOptions()

	// Update the path even if there is no configuration file present. This allows
	// to write to a non existing, non default root (/) location later on.
	cfg.Path = path

	util.Debug.Printf("Reading configuration from: %s\n", path)
	if f, err := os.Open(path); err == nil {
		defer f.Close()

		content, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}

		cfg.Path = path
		cfg, err = parseConfiguration(content, cfg)
		if err != nil {
			return nil, err
		}

		return cfg, nil
	}

	// Initialize collector config with defaults
	SetCollectorConfig(collectors.NewCollectorOptions(map[string]collectorsconfig.CollectorConfig{}))

	return cfg, nil
}

// parseConfiguration parses the YAML configuration file and populates the Options struct
func parseConfiguration(content []byte, cfg *Options) (*Options, error) {
	if len(content) == 0 {
		// Initialize collector config with defaults for empty config files
		SetCollectorConfig(collectors.NewCollectorOptions(map[string]collectorsconfig.CollectorConfig{}))
		return cfg, nil
	}

	var configData configFile
	configData.Options = *cfg

	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)

	if err := decoder.Decode(&configData); err != nil {
		return nil, fmt.Errorf("error parsing configuration: %w", err)
	}

	*cfg = configData.Options

	if err := applyCollectorConfig(configData.Collectors); err != nil {
		return nil, err
	}

	return cfg, nil
}

// applyCollectorConfig validates and applies collector configuration
func applyCollectorConfig(collectorConfigs map[string]collectorConfigEntry) error {
	validatedConfig := make(map[string]collectorsconfig.CollectorConfig)

	for collectorName, entry := range collectorConfigs {
		if entry.Enabled == nil {
			continue
		}

		// Validate collector exists
		if !collectors.IsValidCollector(collectorName) {
			util.Debug.Printf("Warning: Unknown collector '%s' in configuration, skipping\n", collectorName)
			continue
		}

		// Warn if trying to disable a mandatory collector
		if collectors.IsMandatoryCollector(collectorName) && !*entry.Enabled {
			util.Debug.Printf("Warning: Cannot disable mandatory collector '%s', it will remain enabled\n", collectorName)
			continue
		}

		validatedConfig[collectorName] = collectorsconfig.CollectorConfig{
			Enabled: *entry.Enabled,
		}
	}

	SetCollectorConfig(collectors.NewCollectorOptions(validatedConfig))
	return nil
}

// Change the base url to be used when talking to the server to the one being
// provided.
func (opts *Options) ChangeBaseURL(baseUrl string) {
	opts.BaseURL = baseUrl

	// When making an explicit change of the URL, we can further detect which
	// kind of server we are dealing with. For now, let's keep it simple, and if
	// it's the defaultBaseURL then we assume it to be SccProvider, otherwise
	// RmtProvider.
	if opts.BaseURL == defaultBaseURL {
		opts.ServerType = SccProvider
	} else {
		opts.ServerType = RmtProvider
	}
}

// Saves the current options into the configuration path as defined in
// `Options.Path`.
func (opts *Options) SaveAsConfiguration() error {
	buf := bytes.Buffer{}
	fmt.Fprintf(&buf, "---\n")
	fmt.Fprintf(&buf, "url: %s\n", opts.BaseURL)
	fmt.Fprintf(&buf, "insecure: %v\n", opts.Insecure)
	if opts.Language != "" {
		fmt.Fprintf(&buf, "language: %s\n", opts.Language)
	}
	if opts.Namespace != "" {
		fmt.Fprintf(&buf, "namespace: %s\n", opts.Namespace)
	}
	fmt.Fprintf(&buf, "auto_agree_with_licenses: %v\n", opts.AutoAgreeEULA)
	fmt.Fprintf(&buf, "enable_system_uptime_tracking: %v\n", opts.EnableSystemUptimeTracking)

	util.Debug.Printf("Writing configuration to: %s\n", opts.Path)
	return os.WriteFile(opts.Path, buf.Bytes(), 0644)
}

// Returns true if we detected that the configuration points to SCC.
//
// NOTE: this will be reliable if the configuration file already pointed to SCC,
// but it might need to be filled in upon HTTP requests to further guess if it's
// a Glue instance running on localhost or similar developer-only scenarios.
func (opts *Options) IsScc() bool {
	if opts.ServerType == SccProvider {
		return true
	}
	if opts.ServerType == UnknownProvider && opts.BaseURL == defaultBaseURL {
		return true
	}
	return false
}

// Global collector configuration
var collectorConfig collectorsconfig.CollectorOptions = &collectorsconfig.NoCollectorOptions{}

// SetCollectorConfig sets the global collector configuration
func SetCollectorConfig(opts collectorsconfig.CollectorOptions) {
	collectorConfig = opts
}

// GetCollectorConfig returns the current global collector configuration
func GetCollectorConfig() collectorsconfig.CollectorOptions {
	return collectorConfig
}
