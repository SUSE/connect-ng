package collectors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type FakeCollector struct{}

func (collector *FakeCollector) run(arch Architecture) (Result, error) {
	return nil, nil
}

func TestCollectInformationRunAllCollectors(t *testing.T) {
	assert := assert.New(t)
	collectors := []Collector{
		&FakeCollector{},
	}

	_, err := CollectInformation("x86_64", collectors)
	assert.NoError(err)
}
