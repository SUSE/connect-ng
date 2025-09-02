package connect

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/pkg/registration"
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
	BaseURL                    string `json:"url"`
	Language                   string `json:"language"`
	Insecure                   bool   `json:"insecure"`
	Namespace                  string `json:"namespace"`
	FsRoot                     string
	Token                      string
	Product                    registration.Product
	InstanceDataFile           string
	Email                      string `json:"email"`
	AutoAgreeEULA              bool
	EnableSystemUptimeTracking bool
	ServerType                 ServerType
	NoZypperRefresh            bool
	AutoImportRepoKeys         bool
	SkipServiceInstall         bool
	OutputKind                 OutputKind
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

	if f, err := os.Open(path); err == nil {
		defer f.Close()

		cfg.Path = path
		return parseFromConfiguration(f, cfg)
	}
	return cfg, nil
}

// Parse the configuration from the reader `r` and set the corresponding values
// into the given `cfg`. On success it will return the modified configuration,
// otherwise it will return an empty configuration and an error object.
func parseFromConfiguration(r io.Reader, cfg *Options) (*Options, error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if strings.HasPrefix(key, "#") {
			continue
		}
		switch key {
		case "url":
			cfg.BaseURL = val
		case "language":
			cfg.Language = val
		case "namespace":
			cfg.Namespace = val
		case "insecure":
			v, err := strconv.ParseBool(val)
			if err != nil {
				return nil, fmt.Errorf("cannot parse line \"%s\": %v", line, err)
			}
			cfg.Insecure = v
		case "no_zypper_refs":
			v, err := strconv.ParseBool(val)
			if err != nil {
				return nil, fmt.Errorf("cannot parse line \"%s\": %v", line, err)
			}
			cfg.NoZypperRefresh = v
		case "auto_agree_with_licenses":
			v, err := strconv.ParseBool(val)
			if err != nil {
				return nil, fmt.Errorf("cannot parse line \"%s\": %v", line, err)
			}
			cfg.AutoAgreeEULA = v
		case "enable_system_uptime_tracking":
			v, err := strconv.ParseBool(val)
			if err != nil {
				return nil, fmt.Errorf("cannot parse line \"%s\": %v", line, err)
			}
			cfg.EnableSystemUptimeTracking = v
		default:
			return nil, fmt.Errorf("cannot parse line \"%s\" from %s", line, cfg.Path)
		}
	}
	return cfg, nil
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
