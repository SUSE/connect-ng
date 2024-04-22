package collectors

import (
	"strconv"
	"strings"

	"github.com/SUSE/connect-ng/internal/util"
)

type CPU struct{}

func (cpu CPU) run(arch Architecture) (Result, error) {
	output, err := util.Execute([]string{"lscpu", "-p=cpu,socket"}, nil)

	if err != nil {
		return nil, err
	}
	cpus, sockets := parseCPUSocket(strings.TrimSpace(string(output)))

	// We send nil value to SCC to indicate this systems
	// cpu and socket configuration was not available.
	if cpus == 0 || sockets == 0 {
		return Result{"cpus": nil, "sockets": nil}, nil
	}

	return Result{"cpus": cpus, "sockets": sockets}, nil
}

func parseCPUSocket(content string) (int, int) {
	lines := strings.Split(content, "\n")
	last := strings.Split(lines[len(lines)-1], ",")

	cpu, err1 := strconv.Atoi(last[0])
	socket, err2 := strconv.Atoi(last[1])

	if err1 != nil || err2 != nil {
		return 0, 0
	}

	// We take the last line of the lscpu -p=cpu,socket
	// output which is indicates the highest count number
	// of available sockets and cpus but lscpu is 0 indexed
	// Example output:
	/*
		$ lscpu -pcpu,socket
		# The following is the parsable format, which can be fed to other
		# programs. Each different item in every column has an unique ID
		# starting usually from zero.
		# CPU,Socket
		0,0
		1,0
		2,0
		3,0
		4,1
		5,1
		6,1
		7,1
	*/
	return cpu + 1, socket + 1
}
