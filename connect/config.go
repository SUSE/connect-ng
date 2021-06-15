package connect

import (
	"bufio"
	"bytes"
	"fmt"
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
	FsRoot   string
}

func (c Config) toYAML() []byte {
	buf := bytes.Buffer{}
	fmt.Fprintf(&buf, "---\n")
	fmt.Fprintf(&buf, "url: %s\n", c.BaseURL)
	fmt.Fprintf(&buf, "insecure: %v\n", c.Insecure)
	fmt.Fprintf(&buf, "language: %s\n", c.Language)
	return buf.Bytes()
}

// Save saves the config to Path
func (c Config) Save() error {
	data := c.toYAML()
	return os.WriteFile(c.Path, data, 0644)
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
