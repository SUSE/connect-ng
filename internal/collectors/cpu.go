package collectors

import (
	"strconv"
)

type CpuInformation struct {
	Count  int `json:"count"`
	Socket int `json:"socket"`
}

func (cpu CpuInformation) run(arch Architecture) (Result, error) {
	switch arch {
	case ARCHITECTURE_Z:
		return cpuInfoZ()
	default:
		return cpu.cpuInfoDefault()
	}
}

// var cpuCollectorInstance *CpuInformation
// var cpuSingleton sync.Once

// func new() *CpuInformation {
// 	init := func() {
// 		cpuCollectorInstance = &CpuInformation{count: 0, socket: 0}
// 	}
// 	cpuSingleton.Do(init)
// 	return cpuCollectorInstance
// }

func (cpu CpuInformation) cpuInfoDefault() (Result, error) {
	m := make(map[string]interface{})
	var err error
	cpuInfo, err := lscpu()
	if err != nil {
		return nil, err
	}
	cpu.Count, _ = strconv.Atoi(cpuInfo["CPU(s)"])
	cpu.Socket, _ = strconv.Atoi(cpuInfo["Socket(s)"])

	// // DEBUG LOG
	// jsonVal, _ := json.Marshal(cpu)
	// fmt.Println("cpu vlaues calculated: ", string(jsonVal))

	m["cpu"] = cpu
	return m, err
}

func cpuInfoZ() (Result, error) {
	return nil, nil
}

//Expected json structure
// "cpu" :{
// 	"count" : 3
// 	"socket" : 2
// }
