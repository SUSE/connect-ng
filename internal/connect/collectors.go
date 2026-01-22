package connect

import (
	"github.com/SUSE/connect-ng/internal/collectors"
	"github.com/SUSE/connect-ng/pkg/profiles"
)

// Fetch system information based on the available collectors for the system.
func FetchSystemInformation(arch string) (collectors.Result, error) {
	var usedCollectors = []collectors.Collector{
		collectors.CPU{},
		collectors.Hostname{},
		collectors.Memory{},
		collectors.UUID{},
		collectors.Virtualization{},
		collectors.CloudProvider{},
		collectors.Architecture{},
		collectors.ContainerRuntime{},

		// Optional collectors
		collectors.Uname{},
		collectors.SAP{},
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

// Fetch system profile information
func FetchSystemProfiles(arch string, updateCache bool) (collectors.Result, error) {
	var usedCollectors = []collectors.Collector{
		collectors.PCI{UpdateDataIDs: updateCache},
		collectors.LSMOD{UpdateDataIDs: updateCache},
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
