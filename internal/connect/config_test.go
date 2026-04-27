package connect

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var cfg1 = `---
insecure: false
url: https://smt-azure.susecloud.net
language: en_US.UTF-8
no_zypper_refs: true
auto_agree_with_licenses: true
enable_system_uptime_tracking: false`

func TestParseConfig(t *testing.T) {
	assert := assert.New(t)
	content := []byte(cfg1)

	opts := DefaultOptions()
	opts, err := parseConfiguration(content, opts)
	assert.Nil(err)

	expectedURL := "https://smt-azure.susecloud.net"
	assert.Equal(expectedURL, opts.BaseURL)

	expectedLanguage := "en_US.UTF-8"
	assert.Equal(expectedLanguage, opts.Language)

	assert.True(opts.NoZypperRefresh)
	assert.True(opts.AutoAgreeEULA)
}

func TestSaveLoad(t *testing.T) {
	assert := assert.New(t)
	path := filepath.Join(t.TempDir(), "SUSEConnect.test")
	c1 := DefaultOptions()
	c1.Path = path
	c1.AutoAgreeEULA = true
	c1.ServerType = UnknownProvider

	err := c1.SaveAsConfiguration()
	assert.Nil(err)

	c2, err := ReadFromConfiguration(path)
	assert.Nil(err)
	assert.True(reflect.DeepEqual(c1, c2))
}

func TestMinimalAndNonExistingConfiguration(t *testing.T) {
	assert := assert.New(t)

	// Test empty file - should use defaults
	emptyContent := []byte("")
	opts := DefaultOptions()
	result, err := parseConfiguration(emptyContent, opts)
	assert.Nil(err)
	assert.Equal(defaultBaseURL, result.BaseURL)
	assert.Equal(defaultInsecure, result.Insecure)

	// Test minimal YAML file - should use defaults
	minimalContent := []byte("---\n")
	opts2 := DefaultOptions()
	result2, err := parseConfiguration(minimalContent, opts2)
	assert.Nil(err)
	assert.Equal(defaultBaseURL, result2.BaseURL)
	assert.Equal(defaultInsecure, result2.Insecure)

	// Test non-existing file - should use defaults
	nonExistentPath := filepath.Join(t.TempDir(), "does_not_exist")
	result3, err := ReadFromConfiguration(nonExistentPath)
	assert.Nil(err)
	assert.Equal(defaultBaseURL, result3.BaseURL)
	assert.Equal(nonExistentPath, result3.Path)
}

func TestParseValidConfig(t *testing.T) {
	assert := assert.New(t)
	config := `---
url: https://example.com
insecure: true
language: de_DE.UTF-8
namespace: test-namespace
email: user@example.com
auto_agree_with_licenses: true
enable_system_uptime_tracking: true
no_zypper_refs: true`

	opts := DefaultOptions()
	result, err := parseConfiguration([]byte(config), opts)
	assert.Nil(err)

	assert.Equal("https://example.com", result.BaseURL)
	assert.True(result.Insecure)
	assert.Equal("de_DE.UTF-8", result.Language)
	assert.Equal("test-namespace", result.Namespace)
	assert.Equal("user@example.com", result.Email)
	assert.True(result.AutoAgreeEULA)
	assert.True(result.EnableSystemUptimeTracking)
	assert.True(result.NoZypperRefresh)
}

func TestParseInvalidConfigurations(t *testing.T) {
	tests := []struct {
		name   string
		config string
	}{
		{
			name:   "unclosed quote",
			config: `url: "https://example.com`,
		},
		{
			name:   "invalid map structure",
			config: "invalid\n  - list\n  - items",
		},
		{
			name:   "duplicate keys",
			config: "url: https://one.com\nurl: https://two.com",
		},
		{
			name:   "insecure as string",
			config: "insecure: \"not a boolean\"",
		},
		{
			name:   "insecure as number",
			config: "insecure: 123",
		},
		{
			name:   "no_zypper_refs as string",
			config: "no_zypper_refs: \"true\"",
		},
		{
			name:   "unknown field",
			config: "badkey: badval",
		},
		{
			name: "multiple unknown fields",
			config: `---
url: https://example.com
unknown_field: value
another_bad_key: 123`,
		},
		{
			name: "mixed valid and invalid fields",
			config: `---
insecure: true
invalid_option: test
language: en_US`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			opts := DefaultOptions()
			_, err := parseConfiguration([]byte(tt.config), opts)
			assert.NotNil(err)
		})
	}
}

func TestParseConfigWithCollectors(t *testing.T) {
	assert := assert.New(t)
	config := `---
url: https://example.com
collectors:
  pci_devices:
    enabled: false
  sap:
    enabled: true`

	opts := DefaultOptions()
	result, err := parseConfiguration([]byte(config), opts)
	assert.Nil(err)
	assert.Equal("https://example.com", result.BaseURL)

	// Verify collector config was applied
	collectorConfig := GetCollectorConfig()

	assert.False(collectorConfig.IsCollectorEnabled("pci_devices"))
	assert.True(collectorConfig.IsCollectorEnabled("sap"))

	// Verify mandatory collectors are always enabled
	assert.True(collectorConfig.IsCollectorEnabled("cpu"))
}

func TestParseConfigMandatoryCollectorWarning(t *testing.T) {
	assert := assert.New(t)
	config := `---
collectors:
  cpu:
    enabled: false`

	opts := DefaultOptions()
	_, err := parseConfiguration([]byte(config), opts)
	assert.Nil(err)

	// Verify mandatory collector is still enabled despite config
	collectorConfig := GetCollectorConfig()
	assert.True(collectorConfig.IsCollectorEnabled("cpu"))
}

func TestParseConfigUnknownCollector(t *testing.T) {
	assert := assert.New(t)
	config := `---
collectors:
  invalid_collector:
    enabled: true`

	opts := DefaultOptions()
	_, err := parseConfiguration([]byte(config), opts)
	assert.NotNil(err)

	expectedErrMsg := "unknown collector 'invalid_collector' in configuration"
	assert.Equal(expectedErrMsg, err.Error())
}
