package collectors

import (
	"strings"
	"testing"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/assert"
)

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

func mockLscpu(t *testing.T, path string) {
	util.Execute = func(cmd []string, validExitCodes []int) ([]byte, error) {
		actualCmd := strings.Join(cmd, " ")
		testData := util.ReadTestFile(path, t)

		assert.Equal(t, "lscpu -p=cpu,socket", actualCmd, "Wrong command called")

		return testData, nil
	}
}
