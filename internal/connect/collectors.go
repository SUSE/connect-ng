package connect

import (
	"github.com/SUSE/connect-ng/internal/collectors"
	collectorsconfig "github.com/SUSE/connect-ng/pkg/collectors"
	"github.com/SUSE/connect-ng/pkg/profiles"
)

// instantiateCollectors builds a list of collector instances for a given type.
// Handles both direct collectors and factory-based instantiation, with runtime parameters.
func instantiateCollectors(collectorType collectors.CollectorType, updateCache bool, collectorOpts collectorsconfig.CollectorOptions) []collectors.Collector {
	var usedCollectors []collectors.Collector

	for name, entry := range collectors.GetCollectorsByType(collectorType) {
		if collectorOpts.IsCollectorEnabled(name) {
			var c collectors.Collector
			if entry.CollectorFactory != nil {
				c = entry.CollectorFactory(updateCache)
			} else {
				c = entry.Collector
			}
			if c != nil {
				usedCollectors = append(usedCollectors, c)
			}
		}
	}
	return usedCollectors
}

// FetchSystemInformation collects basic system information from enabled collectors.
//
// Parameters:
//   - arch: system architecture (empty string to auto-detect)
//   - collectorOpts: collector configuration options
func FetchSystemInformation(arch string, collectorOpts collectorsconfig.CollectorOptions) (collectors.Result, error) {
	usedCollectors := instantiateCollectors(collectors.SystemInfoCollector, false, collectorOpts)

	var err error
	if arch == "" {
		arch, err = collectors.DetectArchitecture()
	}

	if err != nil {
		return collectors.NoResult, err
	}
	return collectors.CollectInformation(arch, usedCollectors)
}

// FetchSystemProfiles collects system profile data from enabled profile collectors.
//
// Parameters:
//   - arch: system architecture (empty string to auto-detect)
//   - updateCache: whether to update profile data IDs cache
//   - collectorOpts: collector configuration options
func FetchSystemProfiles(arch string, updateCache bool, collectorOpts collectorsconfig.CollectorOptions) (collectors.Result, error) {
	usedCollectors := instantiateCollectors(collectors.ProfileCollector, updateCache, collectorOpts)

	var err error
	if arch == "" {
		arch, err = collectors.DetectArchitecture()
	}

	if err != nil {
		return collectors.NoResult, err
	}
	profile, err := collectors.CollectInformation(arch, usedCollectors)
	if err != nil {
		profiles.DeleteProfileCache("*-profile-id")
		return collectors.NoResult, err
	}
	return profile, err
}
