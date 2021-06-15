package connect

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

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

func (a Config) equals(b Config) bool {
	return a.Path == b.Path && a.BaseURL == b.BaseURL && a.Language == b.Language && a.Insecure == b.Insecure
}

// LoadConfig reads and merges any config from cfgPath if it exists.
// Otherwise just returns config with the default settings.
func LoadConfig(cfgPath string) Config {
	cfg := Config{
		Path:     cfgPath,
		BaseURL:  defaultBaseURL,
		Language: defaultLang,
		Insecure: defaultInsecure,
	}
	f, err := os.Open(cfgPath)
	if err != nil {
		Debug.Println(err)
		cfg.Path = ""
		return cfg
	}
	defer f.Close()
	parseConfig(f, &cfg)
	Debug.Printf("Config after parsing: %+v", cfg)
	return cfg
}

func parseConfig(r io.Reader, cfg *Config) {
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
		case "insecure":
			cfg.Insecure, _ = strconv.ParseBool(val)
		default:
			Debug.Printf("Cannot parse line \"%s\" from %s", line, cfg.Path)
		}
	}
}
