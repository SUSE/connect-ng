package collectors

import (
	"fmt"

	"github.com/SUSE/connect-ng/internal/util"
)

const systemdDetectVirtExecutable = "systemd-detect-virt"

type Virtualization struct{}

func (Virtualization) run(arch string) (Result, error) {
	// Z systems just work differently in this regard, and this is already
	// handled when collecting the CPU info. Hence, there's nothing to do here
	// for these systems.
	if arch == ARCHITECTURE_Z {
		return NoResult, nil
	}

	// We utilize systemd here to fetch the virtualization information. Since we do not hard require systemd
	// there might be the possibility to run on a system without systemd installed (e.g. containers).
	// This tool can detect a lot different hypervisors:
	// https://github.com/systemd/systemd/blob/main/src/basic/virt.c#L1046
	if !util.ExecutableExists(systemdDetectVirtExecutable) {
		return NoResult, fmt.Errorf("can not detect virtualization: `%s` executable not found", systemdDetectVirtExecutable)
	}

	output, err := util.Execute([]string{systemdDetectVirtExecutable, "-v"}, []int{0, 1})

	// systemd-virt-detect does return "none" as output if no virtualization is detected
	if err != nil || string(output) == "none" {
		return NoResult, err
	}

	// Historically SCC API expects virtualization information with the key `hypervisor`. Use it!
	// Check: https://scc.suse.com/connect/v4/documentation#/organizations/put_organizations_systems
	return Result{"hypervisor": string(output)}, nil
}
