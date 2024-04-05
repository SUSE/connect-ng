package collectors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO : rewrite tests
// We're not mocking anything here currently - We were kinda trying to test manually
// This was done so we could actually call the method and see what was being returned
// instead of just mocking the test results

func TestCPUCollectorRun(t *testing.T) {
	assert := assert.New(t)
	expected := make(map[string]interface{})
	// CpuInformation{count: 2, socket: 2} --> result from local run of lscpu
	expected["cpu"] = CpuInformation{Count: 2, Socket: 2}
	testObj := CpuInformation{}
	res, err := testObj.run(ARCHITECTURE_X86_64)
	if err != nil {
		t.Errorf("Something went wrongg: %s", err)
	}

	// if res != expected {
	// 	t.Errorf("Exptected %s \n but got %s", expected, res)
	// }
	assert.Equal(expected, res, "Result mismatch")
}
