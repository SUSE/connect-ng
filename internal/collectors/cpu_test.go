package collectors

import (
	"strings"
	"testing"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/assert"
)

// TODO : rewrite tests
// We're not mocking anything here currently - We were kinda trying to test manually
// This was done so we could actually call the method and see what was being returned
// instead of just mocking the test results

func TestCPUCollectorRun(t *testing.T) {
	assert := assert.New(t)
	mockLscpu(t, "collectors/lscpu_x86_64.txt")
	expected := Result{"cpus": 2, "sockets": 2}
	testObj := CpuInformation{}
	res, err := testObj.run(ARCHITECTURE_X86_64)
	if err != nil {
		t.Errorf("Something went wrongg: %s", err)
	}
	assert.Equal(expected, res, "Result mismatch")
}

// func TestCPUCollectorRunWithInvalidData(t *testing.T) {
// 	assert := assert.New(t)
// 	mockLscpu(t, "collectors/lscpu_x86_64_invalid.txt")
// 	expected := Result{"cpus": 1, "sockets": 1}
// 	testObj := CpuInformation{}
// 	res, err := testObj.run(ARCHITECTURE_X86_64)
// 	if err != nil {
// 		t.Errorf("Something went wrongg: %s", err)
// 	}
// 	assert.Equal(expected, res, "Result mismatch")
// }

func mockLscpu(t *testing.T, path string) {
	util.Execute = func(cmd []string, validExitCodes []int) ([]byte, error) {
		actualCmd := strings.Join(cmd, " ")
		assert.Equal(t, "lscpu", actualCmd, "Wrong command called")
		testData := util.ReadTestFile(path, t)
		return testData, nil
	}
}
