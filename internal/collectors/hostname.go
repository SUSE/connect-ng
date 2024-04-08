package collectors

import "os"

type HostnameInformation struct {
}

func (HostnameInformation) run(arch Architecture) (Result, error) {
	name, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return Result{"hostname": name}, nil
}

// Structure
// {
//   hostname: "gsjdsj"
// }
