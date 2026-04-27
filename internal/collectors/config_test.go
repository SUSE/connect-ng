package collectors

import (
	"testing"

	collectorsconfig "github.com/SUSE/connect-ng/pkg/collectors"
	"github.com/stretchr/testify/assert"
)

func TestNoCollectorOptions(t *testing.T) {
	assert := assert.New(t)
	opts := collectorsconfig.NoCollectorOptions{}

	assert.False(opts.IsCollectorEnabled("any_collector"))
}

func TestGetCollectorsByType(t *testing.T) {
	assert := assert.New(t)
	systemInfoCollectors := GetCollectorsByType(SystemInfoCollector)
	profileCollectors := GetCollectorsByType(ProfileCollector)

	// Verify we have the expected number of each type
	assert.Equal(11, len(systemInfoCollectors)) // 10 mandatory + 1 optional (sap)
	assert.Equal(2, len(profileCollectors))     // pci_devices, kernel_modules

	// Verify specific collectors are in the right category
	_, ok := systemInfoCollectors["cpu"]
	assert.True(ok)

	_, ok = profileCollectors["pci_devices"]
	assert.True(ok)
}

func TestIsMandatoryCollector(t *testing.T) {
	assert := assert.New(t)

	// Mandatory collectors
	mandatoryCollectors := []string{
		"cpu", "hostname", "memory", "uuid", "architecture",
		"virtualization", "cloud_provider", "container_runtime",
		"uname", "vendor",
	}

	for _, name := range mandatoryCollectors {
		assert.True(IsMandatoryCollector(name))
	}

	// Optional collectors
	optionalCollectors := []string{"pci_devices", "kernel_modules", "sap"}

	for _, name := range optionalCollectors {
		assert.False(IsMandatoryCollector(name))
	}

	// Non-existent collector
	assert.False(IsMandatoryCollector("non_existent"))
}

func TestIsValidCollector(t *testing.T) {
	assert := assert.New(t)
	validCollectors := []string{
		"cpu", "hostname", "memory", "uuid", "architecture",
		"virtualization", "cloud_provider", "container_runtime",
		"uname", "vendor", "pci_devices", "kernel_modules", "sap",
	}

	for _, name := range validCollectors {
		assert.True(IsValidCollector(name))
	}

	assert.False(IsValidCollector("invalid_collector"))
}

func TestDefaultCollectorState(t *testing.T) {
	assert := assert.New(t)

	// Enabled by default
	defaultEnabledCollectors := []string{
		"cpu", "hostname", "memory", "uuid", "architecture",
		"virtualization", "cloud_provider", "container_runtime",
		"uname", "vendor", "pci_devices", "kernel_modules",
	}

	for _, name := range defaultEnabledCollectors {
		assert.True(DefaultCollectorState(name))
	}

	// Disabled by default (opt-in)
	assert.False(DefaultCollectorState("sap"))

	// Non-existent collector should default to false
	assert.False(DefaultCollectorState("non_existent"))
}

func TestCollectorOptionsIsCollectorEnabled(t *testing.T) {
	assert := assert.New(t)

	// Test mandatory enforcement
	config := map[string]collectorsconfig.CollectorConfig{
		"cpu": {Enabled: false}, // Try to disable a mandatory collector
	}
	opts := NewCollectorOptions(config)

	assert.True(opts.IsCollectorEnabled("cpu"))

	// Test user configuration for optional collectors
	config2 := map[string]collectorsconfig.CollectorConfig{
		"pci_devices": {Enabled: false},
		"sap":         {Enabled: true},
	}
	opts2 := NewCollectorOptions(config2)

	assert.False(opts2.IsCollectorEnabled("pci_devices"))
	assert.True(opts2.IsCollectorEnabled("sap"))

	// Test default behavior when not in config
	assert.True(opts2.IsCollectorEnabled("kernel_modules"))
}

func TestCollectorRegistryMetadata(t *testing.T) {
	assert := assert.New(t)

	// Verify all mandatory collectors have correct metadata
	for name, entry := range collectorsRegistry {
		if entry.Metadata.Mandatory {
			assert.True(entry.Metadata.DefaultEnabled, "Mandatory collector %s must be default enabled", name)
		}

		// Verify profile collectors have factory functions
		if entry.Metadata.Type == ProfileCollector {
			assert.NotNil(entry.CollectorFactory, "Profile collector %s should have a factory function", name)
		}
	}
}
