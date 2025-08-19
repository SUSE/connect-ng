package connect

import (
	"github.com/SUSE/connect-ng/internal/collectors"
)

// Collectors which are to be used when fetching the system information.
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

// Fetch system information based on the available collectors for the system.
func FetchSystemInformation() (collectors.Result, error) {
	arch, err := collectors.DetectArchitecture()

	if err != nil {
		return collectors.NoResult, err
	}
	return collectors.CollectInformation(arch, usedCollectors)
}
