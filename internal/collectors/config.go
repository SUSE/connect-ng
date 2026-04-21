package collectors

import collectorsconfig "github.com/SUSE/connect-ng/pkg/collectors"

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

// collectorsRegistry is the single source of truth for all collectors
var collectorsRegistry = map[string]CollectorRegistryEntry{
	// Mandatory system information collectors
	"architecture":      {Metadata: CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector}, Collector: Architecture{}},
	"virtualization":    {Metadata: CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector}, Collector: Virtualization{}},
	"cloud_provider":    {Metadata: CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector}, Collector: CloudProvider{}},
	"container_runtime": {Metadata: CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector}, Collector: ContainerRuntime{}},
	"cpu":               {Metadata: CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector}, Collector: CPU{}},
	"memory":            {Metadata: CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector}, Collector: Memory{}},
	"vendor":            {Metadata: CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector}, Collector: Vendor{}},
	"uname":             {Metadata: CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector}, Collector: Uname{}},
	"hostname":          {Metadata: CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector}, Collector: Hostname{}},
	"uuid":              {Metadata: CollectorMetadata{DefaultEnabled: true, Mandatory: true, Type: SystemInfoCollector}, Collector: UUID{}},

	// Optional profile collectors
	"pci_devices": {
		Metadata:         CollectorMetadata{DefaultEnabled: true, Mandatory: false, Type: ProfileCollector},
		CollectorFactory: func(updateDataIDs bool) Collector { return PCI{UpdateDataIDs: updateDataIDs} },
	},
	"kernel_modules": {
		Metadata:         CollectorMetadata{DefaultEnabled: true, Mandatory: false, Type: ProfileCollector},
		CollectorFactory: func(updateDataIDs bool) Collector { return LSMOD{UpdateDataIDs: updateDataIDs} },
	},

	// Optional/specialized collectors
	"sap": {Metadata: CollectorMetadata{DefaultEnabled: false, Mandatory: false, Type: SystemInfoCollector}, Collector: SAP{}},
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
		return config.Enabled
	}

	return DefaultCollectorState(collectorName)
}
