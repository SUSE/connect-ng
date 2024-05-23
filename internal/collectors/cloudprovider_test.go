package collectors

import (
	"strings"
	"testing"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/assert"
)

func mockDmidecodeExists(exists bool) {
	util.ExecutableExists = func(_ string) bool {
		return exists
	}
}

func mockDmidecode(t *testing.T, expectedResult []byte) {
	util.Execute = func(cmd []string, _ []int) ([]byte, error) {
		actualCmd := strings.Join(cmd, " ")
		assert.Equal(t, "dmidecode -t system", actualCmd, "Wrong command called")

		return expectedResult, nil
	}
}

var cloudTestCases = map[string]string{
	"amazon":    "collectors/dmidecode_aws.txt",
	"google":    "collectors/dmidecode_google.txt",
	"microsoft": "collectors/dmidecode_azure.txt",
}

func TestCollectorRunCollectorRun(t *testing.T) {
	assert := assert.New(t)
	testObj := CloudProvider{}

	mockDmidecodeExists(true)

	for expectedProvider, path := range cloudTestCases {
		data := util.ReadTestFile(path, t)
		expected := Result{"cloud_provider": expectedProvider}

		mockDmidecode(t, data)

		result, err := testObj.run(ARCHITECTURE_X86_64)

		assert.NoError(err)
		assert.Equal(expected, result)
	}
}

func TestCloudproviderCollectorRunAWSLarge(t *testing.T) {
	assert := assert.New(t)
	testObj := CloudProvider{}
	data := util.ReadTestFile("collectors/dmidecode_aws_large.txt", t)
	expected := Result{"cloud_provider": "amazon"}

	mockDmidecodeExists(true)
	mockDmidecode(t, data)

	result, err := testObj.run(ARCHITECTURE_X86_64)

	assert.NoError(err)
	assert.Equal(expected, result)
}

func TestCloudproviderCollectorNonCloudEnvironment(t *testing.T) {
	assert := assert.New(t)
	testObj := CloudProvider{}
	data := util.ReadTestFile("collectors/dmidecode_non_cloud.txt", t)

	mockDmidecodeExists(true)
	mockDmidecode(t, data)

	result, err := testObj.run(ARCHITECTURE_X86_64)

	assert.NoError(err)
	assert.Equal(NoResult, result)
}

func TestCloudproviderCollectorDmidecodeNotExisting(t *testing.T) {
	assert := assert.New(t)
	testObj := CloudProvider{}

	mockDmidecodeExists(false)
	result, err := testObj.run(ARCHITECTURE_X86_64)

	assert.Equal(NoResult, result)
	assert.ErrorContains(err, "`dmidecode` executable not found")
}
