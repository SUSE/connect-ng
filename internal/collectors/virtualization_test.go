package collectors

import (
	"strings"
	"testing"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/assert"
)

func mockSystemdDetectVirtExists(exists bool) {
	util.ExecutableExists = func(_ string) bool {
		return exists
	}
}

func mockSystemdDetectVirt(t *testing.T, returnValue string) {
	util.Execute = func(cmd []string, _ []int) ([]byte, error) {
		actualCmd := strings.Join(cmd, " ")

		assert.Equal(t, "systemd-detect-virt -v", actualCmd, "Wrong command called")

		return []byte(returnValue), nil
	}
}

func TestVirtualizationCollectorRun(t *testing.T) {
	assert := assert.New(t)
	testObj := Virtualization{}
	expected := Result{"hypervisor": "kvm"}

	mockSystemdDetectVirtExists(true)
	mockSystemdDetectVirt(t, "kvm")

	result, err := testObj.run(ARCHITECTURE_X86_64)

	assert.NoError(err)
	assert.Equal(expected, result)
}

func TestVirtualizationCollectorRunNoVirtualization(t *testing.T) {
	assert := assert.New(t)
	testObj := Virtualization{}
	expected := Result{"hypervisor": nil}

	mockSystemdDetectVirtExists(true)
	mockSystemdDetectVirt(t, "none")

	result, err := testObj.run(ARCHITECTURE_X86_64)

	assert.NoError(err)
	assert.Equal(expected, result)
}

func TestVirtualizationCollectorRunBinaryMissing(t *testing.T) {
	assert := assert.New(t)
	testObj := Virtualization{}

	mockSystemdDetectVirtExists(false)

	result, err := testObj.run(ARCHITECTURE_X86_64)

	assert.Equal(NoResult, result)
	assert.ErrorContains(err, "can not detect virtualization")
}
