package registration

import "fmt"

// SystemInformation encapsulates any data which should be stored as system meta information.
// This can be sockets, vCPUs, memory or any other useful information.
type SystemInformation = map[string]any

// Indicates there is no system information available when registering.
// Example:
//
//	registration.Register(conn, "hostname", "regcode", NoSystemInformation, NoExtraData)
//
// Note: If possible make sure to include system information whenever possible
var NoSystemInformation = map[string]any{}

// ExtraData encapsulates additional data which might be required to be sent along with the
// request. This can be cloud instance data or online times of the system, depending on the
// requirements.
// Available ExtraData keys are:
//   - instance data (string): Information provided by a cloud instance (Cloud)
//   - namespace (string): Namespace in which the API operates (SMT)
//   - online_at ([]string): OnlineAt definition which records hourly usage time of the system (Cloud)
//   - system_profiles (map[string]any): DataProfiles record of large data profiles
type ExtraData = map[string]any

// Use NoExtraData when no extra information is used by the registration or status method.
// Example:
//
//	registration.Register(conn, "regcode", "hostname", NoSystemInformation, NoExtraData)
var NoExtraData = map[string]any{}

// DataProfiles encapsulates any data which is expected to be big and so it
// needs special encoding and treatment.
type DataProfiles = map[string]any

type requestWithInformation struct {
	SystemInformation any          `json:"hwinfo"`
	InstanceData      string       `json:"instance_data,omitempty"`
	Namespace         string       `json:"namespace,omitempty"`
	OnlineAt          []string     `json:"online_at,omitempty"`
	DataProfiles      DataProfiles `json:"system_profiles,omitempty"`
}

func enrichWithSystemInformation(payload *requestWithInformation, info SystemInformation) {
	payload.SystemInformation = info
}

func enrichWithExtraData(payload *requestWithInformation, extraData ExtraData) error {
	for key, value := range extraData {
		converted := false

		switch key {
		case "instance_data":
			payload.InstanceData, converted = value.(string)
		case "namespace":
			payload.Namespace, converted = value.(string)
		case "online_at":
			payload.OnlineAt, converted = value.([]string)
		case "system_profiles":
			payload.DataProfiles, converted = value.(map[string]any)
		}

		if !converted {
			return fmt.Errorf("Could not parse extra data attribute `%s`. This is a bug.", key)
		}
	}
	return nil
}
