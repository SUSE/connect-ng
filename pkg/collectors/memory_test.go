package collectors

import (
	"bytes"
	"testing"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestMemoryRun(t *testing.T) {
	assert := assert.New(t)
	actual := string(util.ReadTestFile("collectors/meminfo.txt", t))
	mem := Memory{}

	mockLocalOsReadfile(t, "/proc/meminfo", actual)

	result, err := mem.run(ARCHITECTURE_X86_64)
	assert.NoError(err)
	assert.Equal(result["mem_total"], 31766)
}

func TestParseMeminfo(t *testing.T) {
	assert := assert.New(t)
	var tests = []struct {
		file  string
		value int
	}{
		{"MemTotal:       16297236 kB", 15915},
		{"MemTotal:", 0},
		{"MemSomething:       16297236 kB", 0},
		{"Malformed  16297236 kB", 0},
		{"MemTotal:       notanumber kB", 0},
		{"wubalubadubdub", 0},
		{"", 0},
	}

	for _, v := range tests {
		buff := bytes.NewBufferString(v.file)
		val := parseMeminfo(buff)
		assert.Equal(v.value, val)
	}
}
