package collectors

import (
	"fmt"

	"github.com/SUSE/connect-ng/internal/util"
)

type Virtualization struct{}

func (Virtualization) run(arch Architecture) (Result, error) {

	// We utilize systemd here to fetch the virtualization information. Since we do not hard require systemd
	// there might be the possibility to run on a system without systemd installed (e.g. containers).
	// This tool can detect a lot different hypervisors:
	// https://github.com/systemd/systemd/blob/main/src/basic/virt.c#L1046
	if !util.ExecutableExists("systemd-detect-virt") {
		return NoResult, fmt.Errorf("can not detect virtualization: `system-detect-virt` executable not found")
	}

	output, err := util.Execute([]string{"systemd-detect-virt", "-v"}, []int{0, 1})

	if err != nil {
		return NoResult, err
	}

	if string(output) == "none" {
		return Result{"hypervisor": nil}, nil
	}

	// Historicly SCC API expects virtualization information with the key `hypervisor`. Use it!
	// Check: https://scc.suse.com/connect/v4/documentation#/organizations/put_organizations_systems
	return Result{"hypervisor": string(output)}, nil
}
