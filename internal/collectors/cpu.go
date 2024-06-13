package collectors

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/SUSE/connect-ng/internal/util"
)

type CPU struct{}

func (cpu CPU) run(arch string) (Result, error) {
	output, err := util.Execute([]string{"lscpu", "-p=cpu,socket"}, nil)

	if err != nil {
		return nil, err
	}
	cpus, sockets := parseCPUSocket(strings.TrimSpace(string(output)))

	// We send nil value to SCC to indicate this systems
	// cpu and socket configuration was not available.
	if cpus == 0 || sockets == 0 {
		return Result{"cpus": nil, "sockets": nil}, nil
	}

	res := Result{"cpus": cpus, "sockets": sockets}
	return addArchExtras(arch, res), nil
}

func parseCPUSocket(content string) (int, int) {
	lines := strings.Split(content, "\n")
	last := strings.Split(lines[len(lines)-1], ",")

	cpu, err1 := strconv.Atoi(last[0])
	socket, err2 := strconv.Atoi(last[1])

	if err1 != nil || err2 != nil {
		return 0, 0
	}

	// We take the last line of the lscpu -p=cpu,socket
	// output which indicates the highest count number
	// of available sockets and cpus but lscpu is 0 indexed
	// Example output:
	/*
		$ lscpu -pcpu,socket
		# The following is the parsable format, which can be fed to other
		# programs. Each different item in every column has an unique ID
		# starting usually from zero.
		# CPU,Socket
		0,0
		1,0
		2,0
		3,0
		4,1
		5,1
		6,1
		7,1
	*/
	return cpu + 1, socket + 1
}

// Add architecture-specific fields into the given result. Note that this could
// have also been implemented by adding specific `gobuild` tags into
// architecture-specific files, but this would needlessly complicate testing.
func addArchExtras(arch string, result Result) Result {
	if arch == ARCHITECTURE_ARM64 {
		return addArm64Extras(result)
	}
	return result
}

const deviceTreePath = "/sys/firmware/devicetree/base/compatible"

func exactStringMatch(id string, text []byte) string {
	re := regexp.MustCompile(fmt.Sprintf("%v: (.*)", id))
	results := re.FindSubmatch(text)
	if len(results) != 2 {
		return ""
	}
	return string(results[1])
}

// Add extra information that we can gather from an ARM64 machine. This will add
// one extra value to `result`:
//   - `device_tree` (non-ACPI compatible devices): string.
//   - `processor_information (ACPI compatiable devices)`: a map with `family`,
//     `manufacturer` and `signature`.
//
// If nothing could be fetched, then nothing is added and the same `result` is
// returned.
func addArm64Extras(result Result) Result {
	b := util.ReadFile(deviceTreePath)
	if len(b) > 0 {
		// NOTE: the device tree `compatible` file can be weird and contain
		// multiple null bytes spread across the given definition. Hence, `Trim`
		// and friends are not enough and we have to actually replace any
		// occurrences with empty bytes.
		result["device_tree"] = string(bytes.Replace(b, []byte("\x00"), []byte(""), -1))
	} else {
		output, _ := util.Execute([]string{"dmidecode", "-t", "processor"}, nil)

		pinfo := make(map[string]string)
		pinfo["family"] = exactStringMatch("Family", output)
		pinfo["manufacturer"] = exactStringMatch("Manufacturer", output)
		pinfo["signature"] = exactStringMatch("Signature", output)
		if len(pinfo["family"]) == 0 && len(pinfo["manufacturer"]) == 0 && len(pinfo["signature"]) == 0 {
			return result
		}

		result["processor_information"] = pinfo
	}
	return result
}
