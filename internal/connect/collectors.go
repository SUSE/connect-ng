package connect

import (
	"github.com/SUSE/connect-ng/internal/collectors"
)

// Fetch system information based on the available collectors for the system.
func FetchSystemInformation(infoOnly bool) (collectors.Result, error) {
	var usedCollectors = []collectors.Collector{
		collectors.CPU{},
		collectors.Hostname{},
		collectors.Memory{},
		collectors.UUID{},
		collectors.Virtualization{},
		collectors.CloudProvider{},
		collectors.Architecture{},
		collectors.ContainerRuntime{},
		collectors.PCI{UpdateDataIDs: !infoOnly},
		collectors.LSMOD{UpdateDataIDs: !infoOnly},

		// Optional collectors
		collectors.Uname{},
		collectors.SAP{},
	}

	arch, err := collectors.DetectArchitecture()

	if err != nil {
		return collectors.NoResult, err
	}
	return collectors.CollectInformation(arch, usedCollectors)
}
