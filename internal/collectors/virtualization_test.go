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

	mockSystemdDetectVirtExists(true)
	mockSystemdDetectVirt(t, "none")

	result, err := testObj.run(ARCHITECTURE_X86_64)

	assert.NoError(err)
	assert.Equal(NoResult, result)
}

func TestVirtualizationCollectorRunBinaryMissing(t *testing.T) {
	assert := assert.New(t)
	testObj := Virtualization{}

	mockSystemdDetectVirtExists(false)

	result, err := testObj.run(ARCHITECTURE_X86_64)

	assert.Equal(NoResult, result)
	assert.ErrorContains(err, "can not detect virtualization")
}

func TestIsBareMetalPPC(t *testing.T) {
	assert := assert.New(t)

	mockReadFile(t, "/proc/cpuinfo", util.ReadTestFile("collectors/cpuinfo_ppc_bare.txt", t))
	assert.True(isPpcBareMetal())

	mockReadFile(t, "/proc/cpuinfo", util.ReadTestFile("collectors/cpuinfo_ppc_virt.txt", t))
	assert.False(isPpcBareMetal())
}

func TestPowerVM(t *testing.T) {
	assert := assert.New(t)
	testObj := Virtualization{}
	expected := Result{"hypervisor": "powervm"}

	mockReadFile(t, "/proc/cpuinfo", util.ReadTestFile("collectors/cpuinfo_ppc_virt.txt", t))
	mockSystemdDetectVirtExists(true)
	mockSystemdDetectVirt(t, "powervm")

	result, err := testObj.run(ARCHITECTURE_POWER)

	assert.NoError(err)
	assert.Equal(expected, result)
}

func TestPPCLPAR(t *testing.T) {
	assert := assert.New(t)
	testObj := Virtualization{}
	expected := Result{"hypervisor": "lpar"}

	mockReadFile(t, "/proc/cpuinfo", util.ReadTestFile("collectors/cpuinfo_ppc_virt.txt", t))
	mockSystemdDetectVirtExists(true)
	mockSystemdDetectVirt(t, "none")

	result, err := testObj.run(ARCHITECTURE_POWER)

	assert.NoError(err)
	assert.Equal(expected, result)
}
