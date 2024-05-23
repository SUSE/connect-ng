package connect

import (
	"regexp"

	"github.com/SUSE/connect-ng/internal/collectors"
	"github.com/SUSE/connect-ng/internal/util"
)

var (
	cloudEx = `Version: .*(amazon)|Manufacturer: (Amazon)|Manufacturer: (Google)|Manufacturer: (Microsoft) Corporation`
	cloudRe = regexp.MustCompile(cloudEx)
)

const (
	archX86  = "x86_64"
	archARM  = "aarch64"
	archS390 = "s390x"
	archPPC  = "ppc64le"
)

type hwinfo struct {
	Hostname      string `json:"hostname"`
	Cpus          int    `json:"cpus"`
	Sockets       int    `json:"sockets"`
	Clusters      int    `json:"-"`
	Hypervisor    string `json:"hypervisor"`
	Arch          string `json:"arch"`
	UUID          string `json:"uuid"`
	CloudProvider string `json:"cloud_provider"`
	MemTotal      int    `json:"mem_total,omitempty"`
}

func getHwinfo() (hwinfo, error) {
	var err error
	hw := hwinfo{}

	mandatory := []collectors.Collector{
		collectors.CPU{},
		collectors.Hostname{},
		collectors.Memory{},
		collectors.UUID{},
		collectors.Virtualization{},
		collectors.CloudProvider{},
	}

	if hw.Arch, err = arch(); err != nil {
		return hwinfo{}, err
	}

	result, err := collectors.CollectInformation(hw.Arch, mandatory)

	if err != nil {
		return hwinfo{}, err
	}

	// TODO: Handle errors when type casting from interface{}
	hw.Cpus = result["cpus"].(int)
	hw.Sockets = result["sockets"].(int)
	hw.Hostname = result["hostname"].(string)

	switch result["hypervisor"].(type) {
	case string:
		hw.Hypervisor = result["hypervisor"].(string)
	default:
		hw.Hypervisor = ""
	}
	hw.MemTotal = result["mem_total"].(int)
	hw.UUID = result["uuid"].(string) // ignore error to match original

	switch result["cloud_provider"].(type) {
	case string:
		hw.CloudProvider = result["cloud_provider"].(string)
	default:
		hw.CloudProvider = ""
	}

	return hw, nil
}

func arch() (string, error) {
	output, err := util.Execute([]string{"uname", "-i"}, nil)
	if err != nil {
		return "", err
	}
	return string(output), nil
}
