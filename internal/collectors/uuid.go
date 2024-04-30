package collectors

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"

	"github.com/SUSE/connect-ng/internal/util"
)

var localOsReadfile = os.ReadFile

type UUID struct{}

func (UUID) run(arch Architecture) (Result, error) {
	var uuid string
	var err error
	switch arch {
	case ARCHITECTURE_Z:
		uuid, err = uuidS390()
	default:
		uuid, err = uuidDefault()
	}
	if err != nil {
		return NoResult, err
	}

	return Result{"uuid": uuid}, nil
}
func uuidDefault() (string, error) {
	if util.FileExists("/sys/hypervisor/uuid") {
		content, err := localOsReadfile("/sys/hypervisor/uuid")
		if err != nil {
			return "", err
		}
		return string(content), nil
	}
	output, err := util.Execute([]string{"dmidecode", "-s", "system-uuid"}, nil)
	if err != nil {
		return "", err
	}
	out := string(output)
	if strings.Contains(out, "Not Settable") || strings.Contains(out, "Not Present") {
		return "", nil
	}
	return out, nil
}

// uuidS390 returns the system uuid on S390 or "" if it cannot be found
func uuidS390() (string, error) {
	out, err := localOsReadfile("/etc/machine-id")
	if err != nil {
		return "", err
	}

	u, err := uuid.Parse(string(out))
	if err != nil {
		return "", fmt.Errorf("Unable to determine UUID for s390x: %v", err)
	}

	return u.String(), nil
}
