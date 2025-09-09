package collectors

import (
	"github.com/stretchr/testify/mock"
)

type FakeCollector struct {
	mock.Mock
}

func (m *FakeCollector) run(arch string) (Result, error) {
	args := m.Called(arch)

	return args.Get(0).(Result), args.Error(1)
}

func FakeCollectorNew(key string, result any) *FakeCollector {
	obj := FakeCollector{}
	obj.On("run", ARCHITECTURE_X86_64).Return(Result{key: result}, nil)

	return &obj
}
