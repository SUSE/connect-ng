package collectors

import (
	"os"
	"strings"

	"github.com/SUSE/connect-ng/internal/util"
)

type Vendor struct{}

var vendorOsReadfile = os.ReadFile

var deviceTreeModelPaths = []string{
	"/sys/firmware/devicetree/base/model",
	"/sys/firmware/devicetree/hypervisor/model",
}

// try and get the system manufacturer using in order
// dmidecode -s system-manufacture (x86_64 and arm64)
// /sys/firmware/devicetree/base/model
// /sys/firmware/devicetree/hypervisor/model
// Manufacturer value from /procs/ysinfo
func (Vendor) run(arch string) (Result, error) {
	// dmidecode is only on  x86_64 and arm64
	if arch == ARCHITECTURE_X86_64 || arch == ARCHITECTURE_ARM64 {
		output, _ := util.Execute([]string{"dmidecode", "-s", "system-manufacturer"}, nil)
		vendor := strings.TrimSpace(string(output))
		if len(vendor) > 1 {
			return Result{"vendor": string(vendor)}, nil
		}
	}

	model := readDeviceTreeFile(deviceTreeModelPaths)
	if len(model) > 0 {
		return Result{"vendor": strings.TrimSpace(model)}, nil
	}

	sysInfo, _ := vendorOsReadfile("/proc/sysinfo")
	if len(sysInfo) > 0 {
		if vendor := strings.TrimSpace(exactStringMatch("Manufacturer", sysInfo)); vendor != "" {
			return Result{"vendor": string(vendor)}, nil
		}
	}

	return NoResult, nil
}
