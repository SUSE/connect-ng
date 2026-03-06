package collectors

import (
	"strings"

	"github.com/SUSE/connect-ng/internal/util"
)

type Vendor struct{}

// try and get the system manufacturer using in order
// dmidecode -s system-manufacture (x86_64 and arm64
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

	if model, _ := localOsReadfile("/sys/firmware/devicetree/base/model"); len(model) > 0 {
		return Result{"vendor": strings.TrimSpace(string(model))}, nil
	}

	if model := readDeviceTreeFile([]string{"/sys/firmware/devicetree/hypervisor/model"}); len(model) > 0 {
		return Result{"vendor": strings.TrimSpace(model)}, nil
	}
//	if sysInfo := readDeviceTreeFile([]string{"/proc/sysinfo"}); len(sysInfo) > 0 {
	sysInfo, _ := localOsReadfile("/proc/sysinfo")
	if len(sysInfo) > 0 {
		if vendor := strings.TrimSpace(exactStringMatch("Manufacturer", sysInfo)); vendor != "" {
			return Result{"vendor": string(vendor)}, nil
		}
	}

	return NoResult, nil
}
