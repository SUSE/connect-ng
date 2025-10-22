package collectors

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnameCollectorsRun(t *testing.T) {
	assert := assert.New(t)
	testObj := Uname{}

	expectedUname := "6.8.7-1-default #1 SMP PREEMPT_DYNAMIC Thu Apr 18 07:12:38 UTC 2024 (5c0cf23)"
	expectedResult := Result{"uname": expectedUname}

	uname = func(flags []string) (string, error) {
		params := strings.Join(flags, " ")
		if params != "-r -v" {
			assert.Fail("called uname with wrong parameters. Expected `-r -v` but got: `%s`", params)
			return "unknown", nil
		}
		return expectedUname, nil
	}

	result, err := testObj.run(ARCHITECTURE_ARM64)

	assert.NoError(err)
	assert.Equal(expectedResult, result)
}
