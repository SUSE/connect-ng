package collectors

import (
	"testing"

	collectorsconfig "github.com/SUSE/connect-ng/pkg/collectors"
)

func TestNoCollectorOptions(t *testing.T) {
	opts := collectorsconfig.NoCollectorOptions{}

	if opts.IsCollectorEnabled("any_collector") {
		t.Error("NoCollectorOptions should always return false")
	}
}

func TestGetCollectorsByType(t *testing.T) {
	systemInfoCollectors := GetCollectorsByType(SystemInfoCollector)
	profileCollectors := GetCollectorsByType(ProfileCollector)

	// Verify we have the expected number of each type
	if len(systemInfoCollectors) != 11 { // 10 mandatory + 1 optional (sap)
		t.Errorf("Expected 11 system info collectors, got %d", len(systemInfoCollectors))
	}

	if len(profileCollectors) != 2 { // pci_devices, kernel_modules
		t.Errorf("Expected 2 profile collectors, got %d", len(profileCollectors))
	}

	// Verify specific collectors are in the right category
	if _, ok := systemInfoCollectors["cpu"]; !ok {
		t.Error("cpu should be a system info collector")
	}

	if _, ok := profileCollectors["pci_devices"]; !ok {
		t.Error("pci_devices should be a profile collector")
	}
}

func TestIsMandatoryCollector(t *testing.T) {
	// Mandatory collectors
	mandatoryCollectors := []string{
		"cpu", "hostname", "memory", "uuid", "architecture",
		"virtualization", "cloud_provider", "container_runtime",
		"uname", "vendor",
	}

	for _, name := range mandatoryCollectors {
		if !IsMandatoryCollector(name) {
			t.Errorf("%s should be mandatory", name)
		}
	}

	// Optional collectors
	optionalCollectors := []string{"pci_devices", "kernel_modules", "sap"}

	for _, name := range optionalCollectors {
		if IsMandatoryCollector(name) {
			t.Errorf("%s should not be mandatory", name)
		}
	}

	// Non-existent collector
	if IsMandatoryCollector("non_existent") {
		t.Error("non_existent collector should not be mandatory")
	}
}

func TestIsValidCollector(t *testing.T) {
	validCollectors := []string{
		"cpu", "hostname", "memory", "uuid", "architecture",
		"virtualization", "cloud_provider", "container_runtime",
		"uname", "vendor", "pci_devices", "kernel_modules", "sap",
	}

	for _, name := range validCollectors {
		if !IsValidCollector(name) {
			t.Errorf("%s should be a valid collector", name)
		}
	}

	if IsValidCollector("invalid_collector") {
		t.Error("invalid_collector should not be valid")
	}
}

func TestDefaultCollectorState(t *testing.T) {
	// Enabled by default
	defaultEnabledCollectors := []string{
		"cpu", "hostname", "memory", "uuid", "architecture",
		"virtualization", "cloud_provider", "container_runtime",
		"uname", "vendor", "pci_devices", "kernel_modules",
	}

	for _, name := range defaultEnabledCollectors {
		if !DefaultCollectorState(name) {
			t.Errorf("%s should be enabled by default", name)
		}
	}

	// Disabled by default (opt-in)
	if DefaultCollectorState("sap") {
		t.Error("sap should be disabled by default")
	}

	// Non-existent collector should default to false
	if DefaultCollectorState("non_existent") {
		t.Error("non_existent collector should default to false")
	}
}

func TestCollectorOptionsIsCollectorEnabled(t *testing.T) {
	// Test mandatory enforcement
	config := map[string]collectorsconfig.CollectorConfig{
		"cpu": {Enabled: false}, // Try to disable a mandatory collector
	}
	opts := NewCollectorOptions(config)

	if !opts.IsCollectorEnabled("cpu") {
		t.Error("Mandatory collectors should always be enabled")
	}

	// Test user configuration for optional collectors
	config2 := map[string]collectorsconfig.CollectorConfig{
		"pci_devices": {Enabled: false},
		"sap":         {Enabled: true},
	}
	opts2 := NewCollectorOptions(config2)

	if opts2.IsCollectorEnabled("pci_devices") {
		t.Error("pci_devices should be disabled per user config")
	}

	if !opts2.IsCollectorEnabled("sap") {
		t.Error("sap should be enabled per user config")
	}

	// Test default behavior when not in config
	if !opts2.IsCollectorEnabled("kernel_modules") {
		t.Error("kernel_modules should use default (enabled)")
	}
}

func TestCollectorRegistryMetadata(t *testing.T) {
	// Verify all mandatory collectors have correct metadata
	for name, entry := range collectorsRegistry {
		if entry.Metadata.Mandatory {
			if !entry.Metadata.DefaultEnabled {
				t.Errorf("Mandatory collector %s must be default enabled", name)
			}
		}

		// Verify profile collectors have factory functions
		if entry.Metadata.Type == ProfileCollector {
			if entry.CollectorFactory == nil {
				t.Errorf("Profile collector %s should have a factory function", name)
			}
		}
	}
}
