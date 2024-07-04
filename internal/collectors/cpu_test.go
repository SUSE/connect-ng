package collectors

import (
	"errors"
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

func mockReadFileMap(t *testing.T, paths map[string][]byte) {
	util.ReadFile = func(path string) []byte {
		return paths[path]
	}
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

	paths := make(map[string][]byte)
	paths[deviceTreePaths[0]] = util.ReadTestFile("collectors/device_tree_rpi5.txt", t)

	mockReadFileMap(t, paths)
	addArm64Extras(res)

	specs := res["arch_specs"].(map[string]string)
	assert.NotNil(specs)

	assert.Equal("raspberrypi,5-model-bbrcm,bcm2712", specs["device_tree"], "wrong device_tree value")
}

func TestArm64DeviceTreeOtherEntry(t *testing.T) {
	assert := assert.New(t)
	res := Result{}

	paths := make(map[string][]byte)
	paths[deviceTreePaths[0]] = []byte{}
	paths[deviceTreePaths[1]] = util.ReadTestFile("collectors/device_tree_rpi5.txt", t)

	mockReadFileMap(t, paths)
	addArm64Extras(res)

	specs := res["arch_specs"].(map[string]string)
	assert.NotNil(specs)

	assert.Equal("raspberrypi,5-model-bbrcm,bcm2712", specs["device_tree"], "wrong device_tree value")
}

func TestArm64ACPI(t *testing.T) {
	assert := assert.New(t)
	res := Result{}

	paths := make(map[string][]byte)
	paths[deviceTreePaths[0]] = []byte{}

	mockReadFileMap(t, paths)
	mockDmidecode(t, "processor", util.ReadTestFile("collectors/dmidecode_aarch64_acpi.txt", t))
	addArm64Extras(res)

	specs := res["arch_specs"].(map[string]string)
	assert.NotNil(specs)

	assert.Equal("ARMv8", specs["family"], "bad processor family")
	assert.Equal("AppliedMicro(R)", specs["manufacturer"], "bad processor manufacturer")
	assert.Equal(0, len(specs["signature"]), "expecting an empty signature")
}

func TestArm64BadACPI(t *testing.T) {
	assert := assert.New(t)
	res := Result{}

	paths := make(map[string][]byte)
	paths[deviceTreePaths[0]] = []byte{}

	mockReadFileMap(t, paths)
	mockDmidecode(t, "processor", util.ReadTestFile("collectors/dmidecode_aarch64_bad.txt", t))
	addArm64Extras(res)

	assert.Equal(0, len(res), "unexpected result for bad ARM64 ACPI compatible device")
}

func mockReadValuesCmd(path string, t *testing.T) {
	util.Execute = func(cmd []string, _ []int) ([]byte, error) {
		actualCmd := strings.Join(cmd, " ")
		testData := util.ReadTestFile(path, t)

		assert.Equal(t, "read_values -s", actualCmd, "Wrong command called")

		return testData, nil
	}
}

func TestZReadValues(t *testing.T) {
	assert := assert.New(t)

	mockReadValuesCmd("collectors/z_zvm_read_values.txt", t)

	res, err := cpusOnZ()

	assert.NoError(err)

	assert.Equal(2, res["cpus"], t)
	assert.Equal(2, res["sockets"], t)
	assert.Equal("zvm", res["hypervisor"], t)

	specs := res["arch_specs"].(map[string]string)
	assert.Equal("8561", specs["type"], t)
	assert.Equal("ASCHNELL", specs["layer_type"], t)
	_, ok := specs["type_name"]
	assert.False(ok, t)
}

func TestZReadValuesWithName(t *testing.T) {
	assert := assert.New(t)

	mockReadValuesCmd("collectors/z_zvm_read_values_with_type_name.txt", t)

	res, err := cpusOnZ()

	assert.NoError(err)

	assert.Equal(2, res["cpus"], t)
	assert.Equal(2, res["sockets"], t)
	assert.Equal("zvm", res["hypervisor"], t)

	specs := res["arch_specs"].(map[string]string)
	assert.Equal("8561", specs["type"], t)
	assert.Equal("ASCHNELL", specs["layer_type"], t)
	assert.Equal("IBM LinuxONE III", specs["type_name"], t)
}

func TestLPARReadValues(t *testing.T) {
	assert := assert.New(t)

	mockReadValuesCmd("collectors/z_lpar_read_values.txt", t)

	res, err := cpusOnZ()

	assert.NoError(err)
	assert.Equal(6, res["cpus"], t)
	assert.Equal(6, res["sockets"], t)
	assert.Equal("lpar", res["hypervisor"], t)

	specs := res["arch_specs"].(map[string]string)
	assert.Equal("8561", specs["type"], t)
	assert.Equal("ZL01", specs["layer_type"], t)
	_, ok := specs["type_name"]
	assert.False(ok, t)
}

func TestZEmptyReadValues(t *testing.T) {
	assert := assert.New(t)

	mockReadValuesCmd("collectors/empty.txt", t)

	res, err := cpusOnZ()

	assert.NoError(err)
	assert.Nil(res["cpus"], t)
	assert.Nil(res["sockets"], t)
	_, ok := res["hypervisor"]
	assert.False(ok, t)
}

func TestZBadReadValues(t *testing.T) {
	assert := assert.New(t)

	util.Execute = func(cmd []string, _ []int) ([]byte, error) {
		return []byte{}, errors.New("wat")
	}

	res, err := cpusOnZ()
	assert.Nil(res)
	assert.Error(err, "could not execute 'read_values': wat", t)
}

func TestSharedLPAR(t *testing.T) {
	assert := assert.New(t)
	res := Result{}

	paths := make(map[string][]byte)
	paths[deviceTreePaths[0]] = []byte{}
	paths[lparcfgPath] = util.ReadTestFile("collectors/lparcfg_shared.txt", t)

	mockReadFileMap(t, paths)
	addPpc64Extras(res)

	specs := res["arch_specs"].(map[string]string)
	assert.NotNil(specs)
	assert.Equal("shared", specs["lpar_mode"], t)
}

func TestDedicatedLPAR(t *testing.T) {
	assert := assert.New(t)
	res := Result{}

	paths := make(map[string][]byte)
	paths[deviceTreePaths[0]] = []byte{}
	paths[lparcfgPath] = util.ReadTestFile("collectors/lparcfg_dedicated.txt", t)

	mockReadFileMap(t, paths)
	addPpc64Extras(res)

	specs := res["arch_specs"].(map[string]string)
	assert.NotNil(specs)
	assert.Equal("dedicated", specs["lpar_mode"], t)
}
