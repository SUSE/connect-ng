package collectors

import (
	"testing"

	collectorsconfig "github.com/SUSE/connect-ng/pkg/collectors"
	"github.com/stretchr/testify/assert"
)

func TestGetCollectorsByType(t *testing.T) {
	assert := assert.New(t)
	systemInfoCollectors := GetCollectorsByType(SystemInfoCollector)
	profileCollectors := GetCollectorsByType(ProfileCollector)

	// Verify specific collectors are in the right category
	_, ok := systemInfoCollectors["cpu"]
	assert.True(ok)

	_, ok = profileCollectors["pci_data"]
	assert.True(ok)
}

func TestIsMandatoryCollector(t *testing.T) {
	assert := assert.New(t)

	// Test that mandatory collectors in the registry are correctly identified
	for name, entry := range collectorsRegistry {
		if entry.Metadata.Mandatory {
			assert.True(IsMandatoryCollector(name), "collector %s should be mandatory", name)
		} else {
			assert.False(IsMandatoryCollector(name), "collector %s should not be mandatory", name)
		}
	}

	// Non-existent collector should return false
	assert.False(IsMandatoryCollector("non_existent"))
}

func TestIsValidCollector(t *testing.T) {
	assert := assert.New(t)

	// Test that all collectors in the registry are valid
	for name := range collectorsRegistry {
		assert.True(IsValidCollector(name), "collector %s should be valid", name)
	}

	// Test that non-existent collectors are invalid
	assert.False(IsValidCollector("invalid_collector"))
	assert.False(IsValidCollector(""))
}

func TestDefaultCollectorState(t *testing.T) {
	assert := assert.New(t)

	// Test that all collectors in the registry return their configured default state
	for name, entry := range collectorsRegistry {
		expected := entry.Metadata.DefaultEnabled
		actual := DefaultCollectorState(name)
		assert.Equal(expected, actual, "collector %s default state mismatch", name)
	}

	// Non-existent collector should default to false
	assert.False(DefaultCollectorState("non_existent"))
}

func TestCollectorOptionsIsCollectorEnabled(t *testing.T) {
	assert := assert.New(t)

	// Test mandatory enforcement
	config := map[string]collectorsconfig.CollectorConfig{
		"cpu": {State: StateDisabled}, // Try to disable a mandatory collector
	}
	opts := NewCollectorOptions(config)

	assert.True(opts.IsCollectorEnabled("cpu"))

	// Test user configuration for optional collectors
	config2 := map[string]collectorsconfig.CollectorConfig{
		"pci_data": {State: StateDisabled},
	}
	opts2 := NewCollectorOptions(config2)

	assert.False(opts2.IsCollectorEnabled("pci_data"))

	// Test default behavior when not in config
	assert.True(opts2.IsCollectorEnabled("mod_list"))
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

func TestStateValidation(t *testing.T) {
	assert := assert.New(t)

	// Test valid states return their correct enabled status
	validConfig := map[string]collectorsconfig.CollectorConfig{
		"pci_data": {State: StateEnabled},
	}
	opts := NewCollectorOptions(validConfig)
	assert.True(opts.IsCollectorEnabled("pci_data"))

	// Test disabled state
	disabledConfig := map[string]collectorsconfig.CollectorConfig{
		"pci_data": {State: StateDisabled},
	}
	opts2 := NewCollectorOptions(disabledConfig)
	assert.False(opts2.IsCollectorEnabled("pci_data"))

	// Test invalid state falls back to defaults
	invalidConfig := map[string]collectorsconfig.CollectorConfig{
		"pci_data": {State: "invalid_state"},
	}
	opts3 := NewCollectorOptions(invalidConfig)
	assert.True(opts3.IsCollectorEnabled("pci_data")) // Falls back to default (enabled)
}
