package collectors

import (
	"errors"
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectInformationRunAllCollectors(t *testing.T) {
	assert := assert.New(t)

	collector1 := FakeCollectorNew("metric1", "value1")
	collector2 := FakeCollectorNew("metric2", "value2")

	expected := Result{
		"metric1": "value1",
		"metric2": "value2",
	}

	collectors := []Collector{
		collector1,
		collector2,
	}

	result, err := CollectInformation(ARCHITECTURE_X86_64, collectors)
	if assert.NoError(err) {
		assert.Equal(result, expected)
	}
}

func TestCollectInformationWithFailingCollector(t *testing.T) {
	assert := assert.New(t)

	resultCollector1 := Result{"metric1": "value1"}
	resultCollector3 := Result{"metric3": "value3"}

	expectedResult := Result{}
	maps.Copy(expectedResult, resultCollector1)
	maps.Copy(expectedResult, resultCollector3)

	collector1 := FakeCollector{}
	collector2 := FakeCollector{}
	collector3 := FakeCollector{}

	// set up expectations
	collector1.On("run", ARCHITECTURE_X86_64).Return(resultCollector1, nil)
	collector2.On("run", ARCHITECTURE_X86_64).Return(NoResult, errors.New("I am error"))
	collector3.On("run", ARCHITECTURE_X86_64).Return(resultCollector3, nil)

	collectors := []Collector{
		&collector1,
		&collector2,
		&collector3,
	}

	result, err := CollectInformation(ARCHITECTURE_X86_64, collectors)

	assert.NoError(err)
	assert.Equal(expectedResult, result)
}

func TestFromResult(t *testing.T) {
	assert := assert.New(t)
	result := Result{"valueA": 8, "valueB": "some string"}

	assert.Equal(8, FromResult(result, "valueA", 0))
	assert.Equal("some string", FromResult(result, "valueB", "default value"))
	assert.Equal("default value", FromResult(result, "valueC", "default value"))
}
