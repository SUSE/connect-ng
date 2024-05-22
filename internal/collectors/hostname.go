package collectors

import (
	"os"

	"github.com/SUSE/connect-ng/internal/util"
)

type Hostname struct {
}

func (Hostname) run(arch string) (Result, error) {
	name, err := os.Hostname()

	if err != nil || name == "" {
		util.Debug.Printf("Couldn't fetch hostname: %s", err)
		return NoResult, err
	}

	return Result{"hostname": name}, nil
}
