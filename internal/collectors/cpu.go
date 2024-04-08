package collectors

import (
	"strconv"
)

type CpuInformation struct {
}

func (cpu CpuInformation) run(arch Architecture) (Result, error) {
	switch arch {
	case ARCHITECTURE_Z:
		return cpuInfoZ()
	default:
		return cpu.cpuInfoDefault()
	}
}

func (cpu CpuInformation) cpuInfoDefault() (Result, error) {
	cpuInfo, err := lscpu()
	if err != nil {
		return nil, err
	}
	cpuCount, _ := strconv.Atoi(cpuInfo["CPU(s)"])
	socket, _ := strconv.Atoi(cpuInfo["Socket(s)"])

	// // DEBUG LOG
	// jsonVal, _ := json.Marshal(cpu)
	// fmt.Println("cpu vlaues calculated: ", string(jsonVal))

	return Result{"cpus": cpuCount, "sockets": socket}, nil
}

func cpuInfoZ() (Result, error) {
	return nil, nil
}
