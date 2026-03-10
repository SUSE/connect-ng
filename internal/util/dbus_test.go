package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockBusctlExecute(t *testing.T, matcher []string, path string) {
	Execute = func(cmd []string, _ []int) ([]byte, error) {
		testData := ReadTestFile(path, t)

		baseMatcher := []string{BUSCTL_BIN, "--json=short", "--no-pager", "call"}
		baseMatcher = append(baseMatcher, matcher...)

		assert.ElementsMatch(t, cmd, baseMatcher)
		return testData, nil
	}
}

func mockBusctlExecuteFailed(t *testing.T) {
	Execute = func(_ []string, _ []int) ([]byte, error) {
		return []byte{}, fmt.Errorf("busctl call failed")
	}
}

var busctlCommandSample = []string{
	"org.freedesktop.hostname1",
	"/org/freedesktop/hostname1",
	"org.freedesktop.DBus.Properties",
	"Get",
	"ss",
	"org.freedesktop.hostname1",
	"HardwareVendor",
}

func TestBusctlCallSucceeds(t *testing.T) {
	assert := assert.New(t)

	mockBusctlExecute(t, busctlCommandSample, "internal/util/dbus_valid_response.json")

	type dataType []signatureWrapped[string]

	response, err := busctl[dataType](busctlCommandSample...)

	assert.NoError(err)
	assert.Equal(response.Wrapped[0].Wrapped, "Lenovo")
}

func TestBusctlCallCommandFails(t *testing.T) {
	assert := assert.New(t)

	mockBusctlExecuteFailed(t)

	type dataType []signatureWrapped[string]

	response, err := busctl[dataType]()

	assert.Empty(response)
	assert.ErrorContains(err, "busctl call failed")
}

func TestBusctlCallJSONParsingFails(t *testing.T) {
	assert := assert.New(t)

	mockBusctlExecute(t, busctlCommandSample, "internal/util/dbus_valid_response.json")

	// Invalid data type not matching the JSON
	type dataType string

	response, err := busctl[dataType](busctlCommandSample...)

	assert.Empty(response)
	assert.ErrorContains(err, "json: cannot unmarshal array into Go struct field signatureWrapped")
}
