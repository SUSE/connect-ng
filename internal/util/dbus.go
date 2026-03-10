package util

import (
	"encoding/json"
)

// Each JSON encoded response in busctl wraps the data returned.
// The wrapper holds both the DBus signature and interface returned
// data.
// This looks like:
//
//	{
//	  "type": "a{sv}",
//	  "data": [ { "Key": "Value" } ]
//	}
//
// The data structure varies depending on the data returned
// Check https://dbus.freedesktop.org/doc/dbus-specification.html#id-1.3.8
// for more information regarding the provided types
type signatureWrapped[T any] struct {
	Wrapped T `json:"data"`
}

const BUSCTL_BIN = "/usr/bin/busctl"

func busctl[T any](params ...string) (signatureWrapped[T], error) {
	var response signatureWrapped[T]

	command := []string{BUSCTL_BIN, "--json=short", "--no-pager", "call"}
	command = append(command, params...)

	out, err := Execute(command, []int{0})
	if err != nil {
		return response, err
	}

	if err := json.Unmarshal(out, &response); err != nil {
		return response, err
	}

	return response, nil
}
