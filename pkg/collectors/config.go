package collectors

// CollectorConfig holds configuration for a single collector
type CollectorConfig struct {
	Enabled bool
}

// CollectorOptions defines configuration for all data collectors
// This interface allows different systems to provide collector
// configuration without exposing internal implementation details
type CollectorOptions interface {
	// IsCollectorEnabled checks if a collector is enabled
	IsCollectorEnabled(collectorName string) bool
}

// NoCollectorOptions is a default implementation that disables all collectors
// Useful for testing and when collector options are not available
type NoCollectorOptions struct{}

func (NoCollectorOptions) IsCollectorEnabled(collectorName string) bool {
	// Collectors are disabled by default when no configuration is provided
	return false
}
