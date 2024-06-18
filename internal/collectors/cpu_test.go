package collectors

import (
	"strings"
	"testing"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/assert"
)

func mockLscpu(t *testing.T, path string) {
	util.Execute = func(cmd []string, _ []int) ([]byte, error) {
		actualCmd := strings.Join(cmd, " ")
		testData := util.ReadTestFile(path, t)

		assert.Equal(t, "lscpu -p=cpu,socket", actualCmd, "Wrong command called")

		return testData, nil
	}
}

func TestCPUCollectorRun(t *testing.T) {
	assert := assert.New(t)
	expected := Result{"cpus": 8, "sockets": 2}
	testObj := CPU{}

	mockLscpu(t, "collectors/lscpu_x86_64.txt")

	res, err := testObj.run(ARCHITECTURE_X86_64)

	assert.NoError(err)
	assert.Equal(expected, res, "Result mismatch")
}

func TestCPUCollectorRunInvalidCPU(t *testing.T) {
	assert := assert.New(t)
	expected := Result{"cpus": nil, "sockets": nil}
	testObj := CPU{}

	mockLscpu(t, "collectors/lscpu_x86_64_invalid.txt")

	res, err := testObj.run(ARCHITECTURE_X86_64)
	if err != nil {
		t.Errorf("Something went wrong: %s", err)
	}

	assert.NoError(err)
	assert.Equal(expected, res, "Result mismatch")
}

func mockReadFile(t *testing.T, expectedPath string, content []byte) {
	util.ReadFile = func(path string) []byte {
		assert.Equal(t, expectedPath, path)
		return content
	}
}

func TestArm64DeviceTree(t *testing.T) {
	assert := assert.New(t)
	res := Result{}

	mockReadFile(t, deviceTreePath, util.ReadTestFile("collectors/device_tree_rpi5.txt", t))
	addArm64Extras(res)

	assert.Equal("raspberrypi,5-model-bbrcm,bcm2712", res["device_tree"], "wrong device_tree value")
}

func TestArm64ACPI(t *testing.T) {
	assert := assert.New(t)
	res := Result{}

	mockReadFile(t, deviceTreePath, []byte{})
	mockDmidecode(t, "processor", util.ReadTestFile("collectors/dmidecode_aarch64_acpi.txt", t))
	addArm64Extras(res)

	pinfo := res["processor_information"].(map[string]string)
	assert.NotNil(pinfo)

	assert.Equal("ARMv8", pinfo["family"], "bad processor family")
	assert.Equal("AppliedMicro(R)", pinfo["manufacturer"], "bad processor manufacturer")
	assert.Equal(0, len(pinfo["signature"]), "expecting an empty signature")
}

func TestArm64BadACPI(t *testing.T) {
	assert := assert.New(t)
	res := Result{}

	mockReadFile(t, deviceTreePath, []byte{})
	mockDmidecode(t, "processor", util.ReadTestFile("collectors/dmidecode_aarch64_bad.txt", t))
	addArm64Extras(res)

	assert.Equal(0, len(res), "unexpected result for bad ARM64 ACPI compatible device")
}
