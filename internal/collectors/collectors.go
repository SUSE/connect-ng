/*
How to use the collector framework:
In the business logic, we would handle the mandatory/optional metrics
For hw_infos : we can call this framework on mandatory collectors and send the data to scc
For tahoe : we can call this with mandatory collectors. And then specify option collectors too.

Example:

	var MandatoryCollectors = []collectors.Collector{
		collectors.CPU{},
		collectors.Hostname{},
		collectors.Architecture{},
		collectors.Memory{},
		collectors.UUID{},
	}

result, error := collectors.CollectInformation("x86_64", MandatoryCollectors)
*/
package collectors

import (
	"maps"
)

type Result = map[string]interface{}
type Architecture = string

const (
	ARCHITECTURE_X86_64 = "x86_64"
	ARCHITECTURE_ARM64  = "aarch64"
	ARCHITECTURE_POWER  = "ppc64le"
	ARCHITECTURE_Z      = "s390x"
)

var NoResult = Result{}

type Collector interface {
	run(arch Architecture) (Result, error)
}

func CollectInformation(architecture Architecture, collectors []Collector) (Result, error) {
	obj := Result{}

	for _, collector := range collectors {
		res, err := collector.run(architecture)
		if err != nil {
			return NoResult, err
		}
		maps.Copy(obj, res)
	}

	return obj, nil
}
