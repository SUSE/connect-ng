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

func TestGetChkSum(t *testing.T) {
	assert := assert.New(t)
	mockFileExists(true)
	_ = profileCache.PutCacheValue("testPutCheckSum", "testthis")
	tstChkSum := profileCache.GetCacheValue("testPutCheckSum")
	assert.Equal("testthis", tstChkSum)
}

func TestPutChkSum(t *testing.T) {
	assert := assert.New(t)
	err := profileCache.PutCacheValue("testPutCheckSum", "testthis")
	assert.Nil(err)
	tprofileFilePath := GetProfileFilePath()
	content, err1 := os.ReadFile(tprofileFilePath + "testPutCheckSum")
	assert.Nil(err1)
	assert.Equal("testthis", string(content))
}

func TestGetChkSumNoFile(t *testing.T) {
	assert := assert.New(t)
	mockFileExists(false)

	tstChkSum := profileCache.GetCacheValue("/chksum-pcidata.txt")
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
	profileCache.DeleteProfileCache("*")
	files, _ := os.ReadDir(tprofileFilePath)
	assert.Equal(0, len(files))
}

// cleanup dirs created durring test
func TestCleanUp(t *testing.T) {
	_ = os.RemoveAll(GetProfileFilePath())
}

// Following is to test thatthe interface is working as intended
type TestProfileCache struct {
}

var saveCacheVal Profile

var testProfileCache WrappedProfile = &TestProfileCache{}

func (cache *TestProfileCache) DeleteProfileCache(fileFilter string) {
	// dummy method
}

func (cache *TestProfileCache) PutCacheValue(file string, value string) error {
	saveCacheVal.Id = value
	saveCacheVal.Data = file
	return nil
}

func (cache *TestProfileCache) GetCacheValue(file string) string {
	return saveCacheVal.Id
}

// test interface handling
func TestInterfaceHandling(t *testing.T) {
	SetProfileCache(testProfileCache)
	assert := assert.New(t)
	_ = testProfileCache.PutCacheValue("dummy", "works1")
	cacheVal := profileCache.GetCacheValue("dummy")
	assert.Equal("works1", cacheVal)
	_ = testProfileCache.PutCacheValue("dummy", "works3")
	cacheVal = profileCache.GetCacheValue("dummy")
	assert.Equal("works3", cacheVal)
	assert.Equal("dummy", saveCacheVal.Data)
}
