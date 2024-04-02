package collectors

type Result = map[string]interface{}

type Collector interface {
	run() (Result, error)
}

var MandatoryCollectors = []Collector {
	&CpuInformation{},
}

var OptionalCollectors = []Collector {
}

func CollectInformation(architecture string, collectors []Collector) (Result, error) {
	return nil, nil
}

// collectors/cpu.go
type CpuInformation struct {
}
func(cpu *CpuInformation) run() (Result, error) {
	return nil, nil
}
