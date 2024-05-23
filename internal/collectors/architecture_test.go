package collectors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArchitectureCollectorsRun(t *testing.T) {
	assert := assert.New(t)
	testObj := Architecture{}
	expectedResult := Result{"arch": ARCHITECTURE_ARM64}

	result, err := testObj.run(ARCHITECTURE_ARM64)

	assert.NoError(err)
	assert.Equal(expectedResult, result)
}
