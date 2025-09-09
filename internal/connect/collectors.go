package connect

import (
	"github.com/SUSE/connect-ng/internal/collectors"
	"github.com/SUSE/connect-ng/internal/util"
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

	var err error = nil
	if arch == "" {
		arch, err = collectors.DetectArchitecture()
	}

	if err != nil {
		return collectors.NoResult, err
	}
	return collectors.CollectInformation(arch, usedCollectors)
}

// Fetch system profile information
func FetchSystemProfiles(arch string, infoOnly bool) (collectors.Result, error) {
	var usedCollectors = []collectors.Collector{
		collectors.PCI{UpdateDataIDs: !infoOnly},
		collectors.LSMOD{UpdateDataIDs: !infoOnly},
	}

	var err error = nil
	if arch == "" {
		arch, err = collectors.DetectArchitecture()
	}

	if err != nil {
		return collectors.NoResult, err
	}
	profile, err := collectors.CollectInformation(arch, usedCollectors)
	if err != nil {
		util.DeleteProfileCache()
		return collectors.NoResult, err
	}
	return profile, err
}
