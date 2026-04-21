package connect

import (
	"path/filepath"
	"reflect"
	"testing"
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
	if err != nil {
		t.Fatalf("Failed to parse configuration: %v", err)
	}

	expectedURL := "https://smt-azure.susecloud.net"
	if opts.BaseURL != expectedURL {
		t.Fatalf("Unexpected '%v'; expecting '%v'", opts.BaseURL, expectedURL)
	}
	expectedLanguage := "en_US.UTF-8"
	if opts.Language != expectedLanguage {
		t.Fatalf("Unexpected '%v'; expecting '%v'", opts.Language, expectedLanguage)
	}
	if !opts.NoZypperRefresh {
		t.Fatalf("NoZypperRefresh should be true")
	}
	if !opts.AutoAgreeEULA {
		t.Fatalf("AutoAgreeEULA should be true")
	}
}

func TestSaveLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "SUSEConnect.test")
	c1 := DefaultOptions()
	c1.Path = path
	c1.AutoAgreeEULA = true
	c1.ServerType = UnknownProvider
	if err := c1.SaveAsConfiguration(); err != nil {
		t.Fatalf("Unable to write config: %s", err)
	}
	c2, err := ReadFromConfiguration(path)
	if err != nil {
		t.Fatalf("Got an error: %v", err)
	}
	if !reflect.DeepEqual(c1, c2) {
		t.Errorf("got %+v, expected %+v", c2, c1)
	}
}

func TestMinimalAndNonExistingConfiguration(t *testing.T) {
	// Test empty file - should use defaults
	emptyContent := []byte("")
	opts := DefaultOptions()
	result, err := parseConfiguration(emptyContent, opts)
	if err != nil {
		t.Fatalf("Empty configuration should not error, got: %v", err)
	}
	if result.BaseURL != defaultBaseURL {
		t.Errorf("Expected default URL %s, got %s", defaultBaseURL, result.BaseURL)
	}
	if result.Insecure != defaultInsecure {
		t.Errorf("Expected default Insecure %v, got %v", defaultInsecure, result.Insecure)
	}

	// Test minimal YAML file - should use defaults
	minimalContent := []byte("---\n")
	opts2 := DefaultOptions()
	result2, err := parseConfiguration(minimalContent, opts2)
	if err != nil {
		t.Fatalf("Minimal configuration should not error, got: %v", err)
	}
	if result2.BaseURL != defaultBaseURL {
		t.Errorf("Expected default URL %s, got %s", defaultBaseURL, result2.BaseURL)
	}
	if result2.Insecure != defaultInsecure {
		t.Errorf("Expected default Insecure %v, got %v", defaultInsecure, result2.Insecure)
	}

	// Test non-existing file - should use defaults
	nonExistentPath := filepath.Join(t.TempDir(), "does_not_exist")
	result3, err := ReadFromConfiguration(nonExistentPath)
	if err != nil {
		t.Fatalf("Non-existing configuration file should not error, got: %v", err)
	}
	if result3.BaseURL != defaultBaseURL {
		t.Errorf("Expected default URL %s, got %s", defaultBaseURL, result3.BaseURL)
	}
	if result3.Path != nonExistentPath {
		t.Errorf("Expected path %s, got %s", nonExistentPath, result3.Path)
	}
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
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.BaseURL != "https://example.com" {
		t.Errorf("Expected URL https://example.com, got %s", result.BaseURL)
	}
	if !result.Insecure {
		t.Error("Insecure should be true")
	}
	if result.Language != "de_DE.UTF-8" {
		t.Errorf("Expected language de_DE.UTF-8, got %s", result.Language)
	}
	if result.Namespace != "test-namespace" {
		t.Errorf("Expected namespace test-namespace, got %s", result.Namespace)
	}
	if result.Email != "user@example.com" {
		t.Errorf("Expected email user@example.com, got %s", result.Email)
	}
	if !result.AutoAgreeEULA || !result.EnableSystemUptimeTracking || !result.NoZypperRefresh {
		t.Error("All boolean fields should be true")
	}
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
			opts := DefaultOptions()
			_, err := parseConfiguration([]byte(tt.config), opts)
			if err == nil {
				t.Error("Expected error, got nil")
			}
		})
	}
}

func TestParseConfigWithCollectors(t *testing.T) {
	config := `---
url: https://example.com
collectors:
  pci_devices:
    enabled: false
  sap:
    enabled: true`

	opts := DefaultOptions()
	result, err := parseConfiguration([]byte(config), opts)
	if err != nil {
		t.Fatalf("Failed to parse configuration with collectors: %v", err)
	}

	if result.BaseURL != "https://example.com" {
		t.Errorf("Expected URL https://example.com, got %s", result.BaseURL)
	}

	// Verify collector config was applied
	collectorConfig := GetCollectorConfig()

	if collectorConfig.IsCollectorEnabled("pci_devices") {
		t.Error("pci_devices should be disabled per config")
	}

	if !collectorConfig.IsCollectorEnabled("sap") {
		t.Error("sap should be enabled per config")
	}

	// Verify mandatory collectors are always enabled
	if !collectorConfig.IsCollectorEnabled("cpu") {
		t.Error("cpu (mandatory) should always be enabled")
	}
}

func TestParseConfigMandatoryCollectorWarning(t *testing.T) {
	config := `---
collectors:
  cpu:
    enabled: false`

	opts := DefaultOptions()
	_, err := parseConfiguration([]byte(config), opts)
	if err != nil {
		t.Fatalf("Should not error on attempt to disable mandatory collector: %v", err)
	}

	// Verify mandatory collector is still enabled despite config
	collectorConfig := GetCollectorConfig()
	if !collectorConfig.IsCollectorEnabled("cpu") {
		t.Error("cpu (mandatory) should still be enabled even if config tries to disable it")
	}
}

func TestParseConfigUnknownCollector(t *testing.T) {
	config := `---
collectors:
  invalid_collector:
    enabled: true`

	opts := DefaultOptions()
	_, err := parseConfiguration([]byte(config), opts)
	if err == nil {
		t.Fatal("Expected error for unknown collector, got nil")
	}

	expectedErrMsg := "unknown collector 'invalid_collector' in configuration"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}
