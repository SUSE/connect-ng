package util

import (
	"fmt"
)

// The systemd DBus API specification is available as man page:
// https://freedesktop.org/software/systemd/man/latest/org.freedesktop.systemd1.html

type SystemdUnit struct {
	Name        string
	LoadedState string
	ActiveState string
	Path        string
}

func SystemdUnitsByPatterns(patterns []string) ([]SystemdUnit, error) {
	// The returned datatype looks like
	// {
	//   "type": "a(ssssssouso)",
	//   "data": [ [ [ .. tuple (10 elements) .. ] ] ]
	// }
	// Example busctl call:
	// busctl --no-pager --json=pretty call org.freedesktop.systemd1 /org/freedesktop/systemd1 org.freedesktop.systemd1.Manager ListUnitsByPatterns asas 0 1 's*.service'
	type dataType [][][]any

	args := []string{
		"org.freedesktop.systemd1",
		"/org/freedesktop/systemd1",
		"org.freedesktop.systemd1.Manager",
		"ListUnitsByPatterns",
		"asas",
		// all states are considered if no explicit states
		// are supplied
		"0",
		fmt.Sprintf("%d", len(patterns)),
	}
	args = append(args, patterns...)

	response, busErr := busctl[dataType](args...)
	if busErr != nil {
		return []SystemdUnit{}, busErr
	}

	units := []SystemdUnit{}
	for _, data := range response.Wrapped[0] {
		unit := SystemdUnit{
			Name:        forceString(data[0]),
			LoadedState: forceString(data[2]),
			ActiveState: forceString(data[3]),
			Path:        forceString(data[6]),
		}
		units = append(units, unit)
	}

	return units, nil
}

func SystemdServiceBinPath(unit SystemdUnit) (string, error) {
	// Fetching properties of a systemd unit comes with a map like
	// datatype, since each property can utilize its own data structure.
	// It looks like this:
	// {
	//   "type": "v",
	//   "data": [ {
	//       type: "a(sasbttttuii)",
	//       data: [ [ "/bin/path", [ .. args .. ], .. properties ..]
	//     }
	//   ]
	// }
	// Example busctl call:
	// busctl --no-pager --json=pretty call org.freedesktop.systemd1 /org/freedesktop/systemd1/unit/sshd_2eservice org.freedesktop.DBus.Properties Get ss org.freedesktop.systemd1.Service ExecStart
	type dataType []signatureWrapped[[][]any]

	response, busErr := busctl[dataType](
		"org.freedesktop.systemd1",
		unit.Path,
		"org.freedesktop.DBus.Properties",
		"Get",
		"ss",
		"org.freedesktop.systemd1.Service",
		"ExecStart",
	)
	if busErr != nil {
		return "", busErr
	}

	if len(response.Wrapped) > 0 && len(response.Wrapped[0].Wrapped) > 0 {
		unwrapped := response.Wrapped[0].Wrapped[0]

		if len(unwrapped) > 0 {
			return forceString(unwrapped[0]), nil
		}
	}

	return "", fmt.Errorf("can not extract service binary path")
}

func forceString(in any) string {
	switch value := in.(type) {
	case string:
		return value
	// Catch all integers, non-scalar types
	default:
		return fmt.Sprintf("%v", value)
	}
}
