package collectors

import "testing"

type FakeCollector struct {}
func(collector *FakeCollector) run() (Result, error) {
	return nil, nil
}

func TestCollectInformationRunAllCollectors(t *testing.T) {
	collectors := []Collector {
		&FakeCollector{},
	}

	_, err := CollectInformation("x86_64", collectors)
	if err != nil {
		t.Errorf("fail")
	}
}
