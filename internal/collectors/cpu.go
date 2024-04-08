package collectors

import (
	"strconv"
)

type CPU struct {
}

func (cpu CPU) run(arch Architecture) (Result, error) {
	switch arch {
	case ARCHITECTURE_Z:
		return cpuInfoZ()
	default:
		return cpu.cpuInfoDefault()
	}
}

func (CPU) cpuInfoDefault() (Result, error) {
	cpuInfo, err := lscpu()
	if err != nil {
		return NoResult, err
	}

	cpuCount, _ := strconv.Atoi(cpuInfo["CPU(s)"])
	socket, _ := strconv.Atoi(cpuInfo["Socket(s)"])

	return Result{"cpus": cpuCount, "sockets": socket}, nil
}

func cpuInfoZ() (Result, error) {
	return NoResult, nil
}
