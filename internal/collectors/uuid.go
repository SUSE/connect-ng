package collectors

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"

	"github.com/SUSE/connect-ng/internal/util"
)

var localOsReadfile = os.ReadFile

type UUID struct{}

func (UUID) run(arch string) (Result, error) {
	var uuid string
	var err error

	switch arch {
	case ARCHITECTURE_ARM64:
		// ARM machines don't necessarily have `dmidecode` installed (i.e. on
		// non-ACPI devices it does not even make sense to have it installed).
		// Thus, we will first try to fill the uuid if such an executable even
		// exists.
		if util.ExecutableExists("dmidecode") {
			uuid, err = uuidDefault()
		}

		// Either `dmidecode` did not exist, or it returned an empty value. An
		// empty value is also to be expected: on non-ACPI devices that, for
		// whatever reason, have `dmidecode` installed, the tool will simply
		// give back a comment stating that the machine is not supported.
		//
		// In any case, just fallback to using the machine-id value as we do for
		// s390x.
		if err == nil || uuid == "" {
			uuid, err = machineID()
		}
	case ARCHITECTURE_Z:
		uuid, err = machineID()
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

// Returns the UUID as taken from the `/etc/machine-id` file if possible,
// otherwise it returns an empty string.
func machineID() (string, error) {
	out, err := localOsReadfile("/etc/machine-id")
	if err != nil {
		return "", err
	}

	u, err := uuid.Parse(string(bytes.TrimSpace(out)))
	if err != nil {
		return "", fmt.Errorf("unable to determine UUID for s390x: %v", err)
	}

	return u.String(), nil
}
