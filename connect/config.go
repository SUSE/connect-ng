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

// LoadConfig reads and merges any config from cfgPath if it exists.
// Otherwise just returns config with the default settings.
func LoadConfig(cfgPath string) Config {
	c := Config{
		Path:     cfgPath,
		BaseURL:  defaultBaseURL,
		Language: defaultLang,
		Insecure: defaultInsecure,
	}
	f, err := os.Open(cfgPath)
	if err != nil {
		Debug.Println(err)
		c.Path = ""
		return c
	}
	defer f.Close()
	parseConfig(f, &c)
	Debug.Printf("Config after parsing: %+v", c)
	return c
}

func parseConfig(r io.Reader, c *Config) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		idx := strings.Index(line, ":")
		if idx == -1 || len(line) < idx+1 {
			continue
		}
		key, val := line[:idx], line[idx+1:]
		key, val = strings.TrimSpace(key), strings.TrimSpace(val)
		switch key {
		case "url":
			c.BaseURL = val
		case "language":
			c.Language = val
		case "insecure":
			c.Insecure, _ = strconv.ParseBool(val)
		default:
			Debug.Printf("Cannot parse line \"%s\" from %s", line, c.Path)
		}
	}
}
