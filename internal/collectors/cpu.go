package collectors

type CpuInformation struct {
	count int
}

func (cpu *CpuInformation) run(arch Architecture) (Result, error) {
	switch arch {
	case ARCHITECTURE_Z:
		return cpuInfoZ()
	default:
		return cpuInfoDefault()
	}
}

func cpuInfoDefault() (Result, error) {
	m := make(map[string]interface{})
	var err error
	//cpu.count, err = lscpu()
	//m["cpus"] = cpu.count
	return m, err
}

func cpuInfoZ() (Result, error) {
	return nil, nil
}
