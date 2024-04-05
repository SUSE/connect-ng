package collectors

import (
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
	// return m, nil
}

func TestCollectInformationRunAllCollectors(t *testing.T) {
	assert := assert.New(t)
	arch := "x86_64"
	m := Result{"arch": arch}
	expected := map[string]interface{}{"arch": arch}
	testObj := new(FakeCollector)

	// set up expectations
	testObj.On("run", Architecture(arch)).Return(m, nil)

	collectors := []Collector{
		testObj,
	}

	res, err := CollectInformation(Architecture(arch), collectors)
	if assert.NoError(err) {
		assert.Equal(res, expected)
	}
}
