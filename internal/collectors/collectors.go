package collectors

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

var MandatoryCollectors = []Collector{
	&CpuInformation{},
	&HostnameInformation{},
	&ArchitectureInformation{},
	&SocketInformation{},
	&MemoryInformation{},
	&UUIDInformation{},
}

var OptionalCollectors = []Collector{}

func CollectInformation(architecture Architecture, collectors []Collector) (Result, error) {
	return nil, nil
}
