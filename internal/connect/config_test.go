package connect

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var cfg1 = `---
insecure: false
url: https://smt-azure.susecloud.net
language: en_US.UTF-8
no_zypper_refs: true
auto_agree_with_licenses: true
enable_system_uptime_tracking: false`

func TestParseConfig(t *testing.T) {
	content := []byte(cfg1)

	opts := DefaultOptions()
	opts, err := parseConfiguration(content, opts)
	require.NoError(t, err, "Failed to parse configuration")

	assert.Equal(t, "https://smt-azure.susecloud.net", opts.BaseURL)
	assert.Equal(t, "en_US.UTF-8", opts.Language)
	assert.True(t, opts.NoZypperRefresh)
	assert.True(t, opts.AutoAgreeEULA)
}

func TestSaveLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "SUSEConnect.test")
	c1 := DefaultOptions()
	c1.Path = path
	c1.AutoAgreeEULA = true
	c1.ServerType = UnknownProvider
	require.NoError(t, c1.SaveAsConfiguration())

	c2, err := ReadFromConfiguration(path)
	require.NoError(t, err)
	assert.True(t, reflect.DeepEqual(c1, c2))
}

func TestMinimalAndNonExistingConfiguration(t *testing.T) {
	// Test empty file - should use defaults
	emptyContent := []byte("")
	opts := DefaultOptions()
	result, err := parseConfiguration(emptyContent, opts)
	require.NoError(t, err, "Empty configuration should not error")
	assert.Equal(t, defaultBaseURL, result.BaseURL)
	assert.Equal(t, defaultInsecure, result.Insecure)

	// Test minimal YAML file - should use defaults
	minimalContent := []byte("---\n")
	opts2 := DefaultOptions()
	result2, err := parseConfiguration(minimalContent, opts2)
	require.NoError(t, err, "Minimal configuration should not error")
	assert.Equal(t, defaultBaseURL, result2.BaseURL)
	assert.Equal(t, defaultInsecure, result2.Insecure)

	// Test non-existing file - should use defaults
	nonExistentPath := filepath.Join(t.TempDir(), "does_not_exist")
	result3, err := ReadFromConfiguration(nonExistentPath)
	require.NoError(t, err, "Non-existing configuration file should not error")
	assert.Equal(t, defaultBaseURL, result3.BaseURL)
	assert.Equal(t, nonExistentPath, result3.Path)
}

func TestParseValidConfig(t *testing.T) {
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
	require.NoError(t, err)

	assert.Equal(t, "https://example.com", result.BaseURL)
	assert.True(t, result.Insecure)
	assert.Equal(t, "de_DE.UTF-8", result.Language)
	assert.Equal(t, "test-namespace", result.Namespace)
	assert.Equal(t, "user@example.com", result.Email)
	assert.True(t, result.AutoAgreeEULA)
	assert.True(t, result.EnableSystemUptimeTracking)
	assert.True(t, result.NoZypperRefresh)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultOptions()
			_, err := parseConfiguration([]byte(tt.config), opts)
			assert.Error(t, err)
		})
	}
}
