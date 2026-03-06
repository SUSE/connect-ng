package collectors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVendorRunNoModelFound(t *testing.T) {
	assert := assert.New(t)
	vendor := Vendor{}

	mockUtilExecutableExists("dmidecode", false, t)

	result, err := vendor.run(ARCHITECTURE_ARM64)
	assert.NoError(err)
	assert.Equal(NoResult, result)
}

func TestVendorRunDefaultInPhysical(t *testing.T) {
	assert := assert.New(t)
	expectedVendor := "bobs-computer-co"
	vendor := Vendor{}

	mockUtilFileExists(false)
	mockUtilExecute(expectedVendor, nil)

	result, err := vendor.run(ARCHITECTURE_X86_64)
	assert.NoError(err)
	assert.Equal(expectedVendor, result["vendor"])
}

func TestVendorRunSysInfo(t *testing.T) {
	assert := assert.New(t)
//	fileData := "Manufacturer: IBM S390"
	expectedVendor := "IBM S390"
	vendor := Vendor{}

	mockUtilExecute("", nil)

	
//	mockLocalOsReadfile(t, "/proc/sysinfo", "Manufacturer: IBM S390")
	result, err := vendor.run(ARCHITECTURE_Z)
	assert.NoError(err)
	assert.Equal(expectedVendor, result["vendor"])
}
