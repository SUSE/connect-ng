package validation

import "io"

// Enum that contains the possible validation status for a given offline
// activation.
type ValidationStatus int

const (
	// Offline activation is not valid.
	Invalid ValidationStatus = iota

	// Offline activation is valid.
	Valid
)

// Reads activation data from the given reader and validates it by parsing its
// data. Returns a ValidationStatus, the data that is relevant for offline
// activation as a string, and an error if appropiate.
func OfflineActivation(reader io.Reader) (ValidationStatus, string, error) {
	// TODO:
	// 1. Validate the checksum.
	// 2. Parse reader to get *magic*.
	// 3. Return Valid, *magic*, nil

	return Valid, "*magic*", nil
}
