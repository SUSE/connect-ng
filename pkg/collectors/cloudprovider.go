package collectors

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/SUSE/connect-ng/internal/util"
)

var (
	cloudEx = `Version: .*(amazon)|Manufacturer: (Amazon)|Manufacturer: (Google)|Manufacturer: (Microsoft) Corporation`
	cloudRe = regexp.MustCompile(cloudEx)
)

// findCloudProvider returns the cloud provider from "dmidecode -t system" output
func findCloudProvider(b []byte) (string, bool) {
	match := cloudRe.FindSubmatch(b)
	for i, m := range match {
		if i != 0 && len(m) > 0 {
			return string(m), true
		}
	}
	return "", false
}

const dmidecodeExecutable = "dmidecode"

type CloudProvider struct{}

func (CloudProvider) run(_ string) (Result, error) {
	if !util.ExecutableExists(dmidecodeExecutable) {
		return NoResult, fmt.Errorf("can not detect cloud environment: `%s` executable not found", dmidecodeExecutable)
	}

	output, err := util.Execute([]string{dmidecodeExecutable, "-t", "system"}, []int{0})

	if err != nil {
		return NoResult, err
	}

	if provider, found := findCloudProvider(output); found {
		return Result{"cloud_provider": strings.ToLower(provider)}, nil
	}

	return NoResult, nil
}
