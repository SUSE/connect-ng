package collectors

import (
	"testing"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/assert"
)

func mockLocalOsReadfile(t *testing.T, expectedPath string, content string) {
	localOsReadfile = func(path string) ([]byte, error) {
		assert.Equal(t, expectedPath, path)
		return []byte(content), nil
	}
}

func mockUtilFileExists(exists bool) {
	util.FileExists = func(path string) bool {
		return exists
	}
}

func mockUtilExecute(output string, err error) {
	util.Execute = func(_ []string, _ []int) ([]byte, error) {
		return []byte(output), err
	}
}

func TestUUIDRunDefaultInHypervisor(t *testing.T) {
	assert := assert.New(t)
	expectedUUID := "7855822a-6f8b-4dbe-bcc3-e22602d745a9"
	uuid := UUID{}

	mockUtilFileExists(true)
	mockLocalOsReadfile(t, "/sys/hypervisor/uuid", expectedUUID)

	result, err := uuid.run(ARCHITECTURE_X86_64)

	assert.NoError(err)
	assert.Equal(expectedUUID, result["uuid"])
}

func TestUUIDRunDefaultInPhysical(t *testing.T) {
	assert := assert.New(t)
	expectedUUID := "7855822a-6f8b-4dbe-bcc3-e22602d745a9"
	uuid := UUID{}

	mockUtilFileExists(false)
	mockUtilExecute(expectedUUID, nil)

	result, err := uuid.run(ARCHITECTURE_X86_64)
	assert.NoError(err)
	assert.Equal(expectedUUID, result["uuid"])
}

func TestUUIDRunS390x(t *testing.T) {
	assert := assert.New(t)
	actualUUID := "a85d8326f09347ef9f118da1a74a4dd1"
	expectedUUID := "a85d8326-f093-47ef-9f11-8da1a74a4dd1"
	uuid := UUID{}

	mockLocalOsReadfile(t, "/etc/machine-id", actualUUID)

	result, err := uuid.run(ARCHITECTURE_Z)
	assert.NoError(err)
	assert.Equal(expectedUUID, result["uuid"])
}
