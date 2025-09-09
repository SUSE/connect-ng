package profiles

import (
	"os"
	"testing"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/assert"
)

func mockFileExists(exists bool) {
	util.FileExists = func(path string) bool {
		return exists
	}
}

func TestSetProfilePath(t *testing.T) {
	assert := assert.New(t)
	expected, _ := os.MkdirTemp("/tmp/", "__suseconnect")
	SetProfileFilePath(expected)
	actual := GetProfileFilePath()
	assert.Equal(expected+"/", actual)
}

func TestPutChkSum(t *testing.T) {
	assert := assert.New(t)
	err := PutCacheValue("testPutCheckSum", "testthis")
	assert.Nil(err)
	tprofileFilePath := GetProfileFilePath()
	content, err1 := os.ReadFile(tprofileFilePath + "testPutCheckSum")
	assert.Nil(err1)
	assert.Equal("testthis", string(content))
}

func TestGetChkSum(t *testing.T) {
	assert := assert.New(t)
	mockFileExists(true)
	_ = PutCacheValue("testPutCheckSum", "testthis")
	tstChkSum := GetCacheValue("testPutCheckSum")
	assert.Equal("testthis", tstChkSum)
}

func TestGetChkSumNoFile(t *testing.T) {
	assert := assert.New(t)
	mockFileExists(false)

	tstChkSum := GetCacheValue("chksum-pcidata.txt")
	assert.Equal("", tstChkSum)
}

func TestCalcSha256(t *testing.T) {
	assert := assert.New(t)

	actual := calcSha256("623fgauib7v onoacn'vm4nsMv9wn34mcpmqw35cp;54m")
	expected := "7e8871c2ffafe97edfd6e90226adf466d07c7f64a8f56f51b25a74cb34e2b688"
	assert.Equal(expected, actual)
}

func TestDeleteProfileCache(t *testing.T) {
	assert := assert.New(t)
	tprofileFilePath := GetProfileFilePath()
	DeleteProfileCache()
	_, err := os.Stat(tprofileFilePath)
	assert.NotNil(err)
}
