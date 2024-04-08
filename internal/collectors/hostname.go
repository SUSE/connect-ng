package collectors

import "os"

type Hostname struct {
}

func (Hostname) run(arch Architecture) (Result, error) {
	name, err := os.Hostname()
	if err != nil {
		return NoResult, err
	}

	return Result{"hostname": name}, nil
}
