package collectors

import (
	"maps"
)

type Result = map[string]interface{}
type Architecture string

const (
	ARCHITECTURE_X86_64 = "x86_64"
	ARCHITECTURE_ARM64  = "aarch64"
	ARCHITECTURE_POWER  = "ppc64le"
	ARCHITECTURE_Z      = "s390x"
)

type Collector interface {
	run(arch Architecture) (Result, error)
}

// type Collectable struct {
// 	_json_key string
// 	_json_val Result
// }
// type ICollect struct {
// 	Collector
// 	Collectable
// }

// func (c *Collectable) get_key() string {
// 	return c._json_key
// }

// func (c *Collectable) UnmarshalJSON(data []byte) error {
// 	var m map[string]interface{}
// 	if err := json.Unmarshal(data, &m); err != nil {
// 		return err
// 	}
// 	c._json_key = c.get_key()
// 	c._json_val = m
// 	return nil
// }

// func (c *Collectable) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(&struct {
// 		Arch string `json:"arch"`
// 	}{
// 		Arch: c._json_val[c.get_key()].(string),
// 	})
// }

var MandatoryCollectors = []Collector{
	// &CpuInformation{},
	// &HostnameInformation{},
	&ArchitectureInformation{},
	// &SocketInformation{},
	// &MemoryInformation{},
	// &UUIDInformation{},
}

var OptionalCollectors = []Collector{}

func CollectInformation(architecture Architecture, collectors []Collector) (Result, error) {
	obj := make(Result)
	for _, collector := range collectors {
		res, err := collector.run(architecture)
		if err != nil {
			return nil, err
		}
		// obj[collector.get_key()] = res[collector.get_key()]
		maps.Copy(obj, res)
	}
	return obj, nil
}
