package connect

import (
	"github.com/SUSE/connect-ng/internal/collectors"
	"github.com/SUSE/connect-ng/pkg/profiles"
)

// FetchSystemInformation collects basic system information from enabled collectors.
//
// Parameters:
//   - arch: system architecture (empty string to auto-detect)
func FetchSystemInformation(arch string) (collectors.Result, error) {
	collectorOpts := GetCollectorConfig()

	var usedCollectors []collectors.Collector

	for name, entry := range collectors.GetCollectorsByType(collectors.SystemInfoCollector) {
		if collectorOpts.IsCollectorEnabled(name) {
			usedCollectors = append(usedCollectors, entry.Collector)
		}
	}

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
func FetchSystemProfiles(arch string, updateCache bool) (collectors.Result, error) {
	collectorOpts := GetCollectorConfig()

	var usedCollectors []collectors.Collector

	for name, entry := range collectors.GetCollectorsByType(collectors.ProfileCollector) {
		if collectorOpts.IsCollectorEnabled(name) {
			// Use factory function to create collector with runtime parameter
			if entry.CollectorFactory != nil {
				usedCollectors = append(usedCollectors, entry.CollectorFactory(updateCache))
			} else {
				usedCollectors = append(usedCollectors, entry.Collector)
			}
		}
	}

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
