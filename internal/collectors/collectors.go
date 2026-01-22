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

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/pkg/profiles"
)

type Result = profiles.Result

const (
	ARCHITECTURE_X86_64 = "x86_64"
	ARCHITECTURE_ARM64  = "aarch64"
	ARCHITECTURE_POWER  = "ppc64le"
	ARCHITECTURE_Z      = "s390x"
)

var NoResult = Result{}

type Collector interface {
	run(arch string) (Result, error)
}

func CollectInformation(architecture string, collectors []Collector) (Result, error) {
	obj := Result{}

	for _, collector := range collectors {
		res, err := collector.run(architecture)
		if err != nil {
			util.Debug.Printf("Collecting system information failed: %s", err)
		}
		maps.Copy(obj, res)
	}

	return obj, nil
}

// Extract a value from the already existing result set preserving the existing value type
// and providing a default value in case the to be extracted key does not existing in the
// result set or is of different type
func FromResult[R any](result Result, key string, def R) R {
	if value, ok := result[key]; ok {
		switch value.(type) {
		case R:
			return value.(R)
		default:
			return def
		}
	}
	return def
}
