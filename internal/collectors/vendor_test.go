package collectors

import (
	"testing"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/assert"
)

// mockVendorOsReadfile
// takes file path and contents to return when file "read"
//
//	effects all uses of vendorOsReadfile
//	only supports mocking one file at a time
//	if passed file is not being mocked return nil, nil
func mockVendorOsReadfile(t *testing.T, expectedPath string, content string) {
	vendorOsReadfile = func(path string) ([]byte, error) {
		if expectedPath == path {
			return []byte(content), nil
		}
		return nil, nil
	}
}

// mockVendorUtilReadfile
// takes file path and contents to return when file "read"
//
//	effects all uses of util.ReadFile
//	only supports mocking one file at a time
//	if passed file is not being mocked return nil, nil
func mockVendorUtilReadfile(t *testing.T, expectedPath string, content []byte) {
	util.ReadFile = func(path string) []byte {
		if expectedPath == path {
			return content
		}
		return nil
	}
}

func TestVendorRunNoModelFound(t *testing.T) {
	assert := assert.New(t)
	vendor := Vendor{}

	// block getting vendor from local FS
	mockVendorOsReadfile(t, "/no/file/here", "")
	mockVendorUtilReadfile(t, "/no/file/here", nil)

	result, err := vendor.run(ARCHITECTURE_ARM64)
	assert.NoError(err)
	assert.Equal(NoResult, result)
}

func TestVendorRunDefaultInPhysical(t *testing.T) {
	assert := assert.New(t)
	expectedVendor := "bobs-computer-co"
	vendor := Vendor{}

	// block getting vendor from local FS
	mockVendorOsReadfile(t, "/no/file/here", "")
	mockVendorUtilReadfile(t, "/no/file/here", nil)
	mockUtilExecute(expectedVendor, nil)

	result, err := vendor.run(ARCHITECTURE_X86_64)
	assert.NoError(err)
	assert.Equal(expectedVendor, result["vendor"])
}

func TestVendorBaseSysInfo(t *testing.T) {
	assert := assert.New(t)
	fileData := "Base Test"
	expectedVendor := "Base Test"
	vendor := Vendor{}

	mockUtilExecute("", nil)
	mockVendorOsReadfile(t, "/no/file/here", "")
	mockVendorUtilReadfile(t, "/sys/firmware/devicetree/base/model", []byte(fileData))
	result, err := vendor.run(ARCHITECTURE_ARM64)
	assert.NoError(err)
	assert.Equal(expectedVendor, result["vendor"])
}

func TestVendoHypervModel(t *testing.T) {
	assert := assert.New(t)
	fileData := "HyperV"
	expectedVendor := "HyperV"
	vendor := Vendor{}

	mockVendorOsReadfile(t, "/no/file/here", "")
	mockVendorUtilReadfile(t, "/sys/firmware/devicetree/hypervisor/model", []byte(fileData))
	result, err := vendor.run(ARCHITECTURE_Z)
	assert.NoError(err)
	assert.Equal(expectedVendor, result["vendor"])
}

func TestVendorProcSysInfo(t *testing.T) {
	assert := assert.New(t)
	fileData := "Manufacturer: IBM S390"
	expectedVendor := "IBM S390"
	vendor := Vendor{}

	mockVendorUtilReadfile(t, "/sys/firmware/devicetree/hypervisor/model", nil)
	mockVendorOsReadfile(t, "/proc/sysinfo", fileData)
	result, err := vendor.run(ARCHITECTURE_Z)
	assert.NoError(err)
	assert.Equal(expectedVendor, result["vendor"])
}
