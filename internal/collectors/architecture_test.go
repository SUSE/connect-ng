package collectors

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockArchitectureInformation struct {
	mock.Mock
}

func (m *MockArchitectureInformation) run(arch Architecture) (Result, error) {
	args := m.Called(arch)

	return args.Get(0).(Result), args.Error(1)
	// return m, nil
}

func TestArchitectureCollectorRun(t *testing.T) {
	arch := "x86_64"
	m := Result{"arch": arch}
	testObj := new(MockArchitectureInformation)

	// set up expectations
	testObj.On("run", Architecture(arch)).Return(m, nil)

	// call the code we are testing
	testingFunc := func(t *testing.T, arch Architecture, colls []Collector) {
		for _, col := range colls {
			_, err := col.run(arch)
			if err != nil {
				t.Fail()
			}
		}
	}
	testingFunc(t, Architecture(arch), []Collector{testObj})
	testObj.AssertExpectations(t)

}
