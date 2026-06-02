package collectors

import collectorsconfig "github.com/SUSE/connect-ng/pkg/collectors"

// Collector state constants
const (
	StateEnabled  = "enabled"
	StateDisabled = "disabled"
	// In the future we might want to track states such as:
	// OptIn  = "opt-in"
	// OptOut = "opt-out"
)

// stateToEnabled maps collector states to their enabled status
// When adding new states, add them here with their enabled value
var stateToEnabled = map[string]bool{
	StateEnabled:  true,
	StateDisabled: false,
	// Future states:
	// OptIn:  true,   // opt-in means enabled
	// OptOut: false,  // opt-out means disabled
}

// CollectorType distinguishes between system information and profile collectors
type CollectorType int

const (
	SystemInfoCollector CollectorType = iota
	ProfileCollector
)

// CollectorMetadata holds metadata about a collector
type CollectorMetadata struct {
	DefaultEnabled bool
	Mandatory      bool
	Type           CollectorType
}

// CollectorRegistryEntry represents a single collector's registry entry
type CollectorRegistryEntry struct {
	Metadata         CollectorMetadata
	Collector        Collector
	CollectorFactory func(updateDataIDs bool) Collector
}

// collectorsRegistry is the single source of truth for all collectors.
// All collector types referenced here are defined in this package.
// When adding a new collector: define the collector type in this package,
// then add an entry here with appropriate metadata.
var collectorsRegistry = map[string]CollectorRegistryEntry{
	// Mandatory collectors
	"arch": {
		Metadata:  CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector},
		Collector: Architecture{},
	},
	"hypervisor": {
		Metadata:  CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector},
		Collector: Virtualization{},
	},
	"cloud_provider": {
		Metadata:  CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector},
		Collector: CloudProvider{},
	},
	"container_runtime": {
		Metadata:  CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector},
		Collector: ContainerRuntime{},
	},
	"cpus": {
		Metadata:  CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector},
		Collector: CPU{},
	},
	"mem_total": {
		Metadata:  CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector},
		Collector: Memory{},
	},
	"vendor": {
		Metadata:  CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector},
		Collector: Vendor{},
	},
	"uname": {
		Metadata:  CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector},
		Collector: Uname{},
	},
	"hostname": {
		Metadata:  CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector},
		Collector: Hostname{},
	},
	"uuid": {
		Metadata:  CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector},
		Collector: UUID{},
	},
	"sap": {
		Metadata:  CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector},
		Collector: SAP{},
	},
	"kubernetes_provider": {
		Metadata:  CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector},
		Collector: K8S{},
	},
	"ha_active": {
		Metadata:  CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector},
		Collector: HA{},
	},

	// Optional collectors
	"pci_data": {
		Metadata:         CollectorMetadata{DefaultEnabled: true, Mandatory: false, Type: ProfileCollector},
		CollectorFactory: func(updateDataIDs bool) Collector { return PCI{UpdateDataIDs: updateDataIDs} },
	},
	"mod_list": {
		Metadata:         CollectorMetadata{DefaultEnabled: true, Mandatory: false, Type: ProfileCollector},
		CollectorFactory: func(updateDataIDs bool) Collector { return LSMOD{UpdateDataIDs: updateDataIDs} },
	},
	"installed_pkgs": {
		Metadata:         CollectorMetadata{DefaultEnabled: true, Mandatory: false, Type: ProfileCollector},
		CollectorFactory: func(updateDataIDs bool) Collector { return InstalledPackages{UpdateDataIDs: updateDataIDs} },
	},
}

// GetCollectorsByType returns all collectors of a specific type
func GetCollectorsByType(collectorType CollectorType) map[string]CollectorRegistryEntry {
	result := make(map[string]CollectorRegistryEntry)
	for name, entry := range collectorsRegistry {
		if entry.Metadata.Type == collectorType {
			result[name] = entry
		}
	}
	return result
}

// IsMandatoryCollector checks if a collector is mandatory
func IsMandatoryCollector(collectorName string) bool {
	if entry, ok := collectorsRegistry[collectorName]; ok {
		return entry.Metadata.Mandatory
	}
	return false
}

// IsValidCollector checks if a collector name exists in the registry
func IsValidCollector(collectorName string) bool {
	_, ok := collectorsRegistry[collectorName]
	return ok
}

// DefaultCollectorState returns the default enabled state for a collector
func DefaultCollectorState(collectorName string) bool {
	if entry, ok := collectorsRegistry[collectorName]; ok {
		return entry.Metadata.DefaultEnabled
	}
	return false
}

type CollectorOptions struct {
	collectors map[string]collectorsconfig.CollectorConfig
}

func NewCollectorOptions(collectors map[string]collectorsconfig.CollectorConfig) *CollectorOptions {
	return &CollectorOptions{
		collectors: collectors,
	}
}

func (c *CollectorOptions) IsCollectorEnabled(collectorName string) bool {
	if IsMandatoryCollector(collectorName) {
		return true
	}

	if config, ok := c.collectors[collectorName]; ok {
		if enabled, validState := stateToEnabled[config.State]; validState {
			return enabled
		}
	}

	return DefaultCollectorState(collectorName)
}
