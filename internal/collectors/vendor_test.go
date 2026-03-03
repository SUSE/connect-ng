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
