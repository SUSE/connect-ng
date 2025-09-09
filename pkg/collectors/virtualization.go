package collectors

import (
	"fmt"
	"regexp"

	"github.com/SUSE/connect-ng/internal/util"
)

const systemdDetectVirtExecutable = "systemd-detect-virt"

type Virtualization struct{}

func (Virtualization) run(arch string) (Result, error) {
	if arch == ARCHITECTURE_Z {
		// Z systems just work differently in this regard, and this is already
		// handled when collecting the CPU info. Hence, there's nothing to do
		// here for these systems.
		return NoResult, nil
	} else if arch == ARCHITECTURE_POWER && isPpcBareMetal() {
		// We can detect early on if we are on bare metal PowerPC.
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

	// If there was any trouble with the executable, return a no result. This is
	// hardly going to happen given that we checked for the existence of the
	// executable beforehand.
	if err != nil {
		return NoResult, err
	}

	// systemd-virt-detect returns "none" as output if no virtualization was
	// detected from their side.
	if string(output) == "none" {
		// That being said, if we are on PowerPC and we did not detect it as
		// bare metal before, then we are most probably on an LPAR environment.
		if arch == ARCHITECTURE_POWER {
			return Result{"hypervisor": "lpar"}, nil
		}
		return NoResult, err
	}

	// Historically SCC API expects virtualization information with the key `hypervisor`. Use it!
	// Check: https://scc.suse.com/connect/v4/documentation#/organizations/put_organizations_systems
	return Result{"hypervisor": string(output)}, nil
}

var ppcRe = regexp.MustCompile(`platform\s*:\s*pSeries`)

// Returns if the given PowerPC machine is running on bare metal or not.
func isPpcBareMetal() bool {
	// Implementation taken from the hints from:
	// http://git.annexia.org/?p=virt-what.git;a=blob;f=virt-what.in;h=5ccf49e2e7a584ac4ead259c903026ece510fee9;hb=HEAD#l423

	s := util.ReadFile("/proc/cpuinfo")
	if len(s) == 0 {
		return false
	}
	results := ppcRe.FindSubmatch(s)
	return len(results) != 1
}
