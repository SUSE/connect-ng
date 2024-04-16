package collectors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type FakeCollector struct {
	mock.Mock
}

func (m *FakeCollector) run(arch Architecture) (Result, error) {
	args := m.Called(arch)

	return args.Get(0).(Result), args.Error(1)
}

func TestCollectInformationRunAllCollectors(t *testing.T) {
	assert := assert.New(t)

	resultCollector1 := Result{"metric1": "value1"}
	resultCollector2 := Result{"metric2": "value2"}

	collector1 := FakeCollector{}
	collector2 := FakeCollector{}

	expected := Result{
		"metric1": "value1",
		"metric2": "value2",
	}

	// set up expectations
	collector1.On("run", ARCHITECTURE_X86_64).Return(resultCollector1, nil)
	collector2.On("run", ARCHITECTURE_X86_64).Return(resultCollector2, nil)

	collectors := []Collector{
		&collector1,
		&collector2,
	}

	result, err := CollectInformation(ARCHITECTURE_X86_64, collectors)
	if assert.NoError(err) {
		assert.Equal(result, expected)
	}
}

func TestCollectInformationWithFailingCollector(t *testing.T) {
	assert := assert.New(t)

	resultCollector1 := Result{"metric1": "value1"}

	collector1 := FakeCollector{}
	collector2 := FakeCollector{}

	// set up expectations
	collector1.On("run", ARCHITECTURE_X86_64).Return(resultCollector1, nil)
	collector2.On("run", ARCHITECTURE_X86_64).Return(NoResult, errors.New("I am error"))

	collectors := []Collector{
		&collector1,
		&collector2,
	}

	result, err := CollectInformation(ARCHITECTURE_X86_64, collectors)
	assert.Error(err)
	assert.Equal(result, NoResult)
}
