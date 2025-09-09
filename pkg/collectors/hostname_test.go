package collectors

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostnameRun(t *testing.T) {
	assert := assert.New(t)
	hostName, _ := os.Hostname()
	expected := Result{"hostname": hostName}
	collector := Hostname{}

	result, err := collector.run(ARCHITECTURE_X86_64)

	assert.Equal(expected, result)
	assert.Nil(err)
}
