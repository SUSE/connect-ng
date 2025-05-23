package testutil

import (
	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/stretchr/testify/mock"
)

type MockWrappedAPI struct {
	mock.Mock
}

// Use this to mock any API interaction in SUSEConnect.
// Example:
//
//	api := NewMockWrappedAPI()
//	expected := []labels.Label{
//	  []labels.Label{name: "label-1"},
//	}
//
//	api.On("AssignLabels").Return(labels, nil)
//	result, err := MethodWhichAssignsLabels(api, []string{"label-1",})
//	assert.NoError(err)
//	assert.Equal(expected, result)
func NewMockWrappedAPI() *MockWrappedAPI {
	api := &MockWrappedAPI{}

	return api
}

// INFO: For information how to handle arguments in the mocked interface implementation
// check: https://pkg.go.dev/github.com/stretchr/testify/mock

func (m *MockWrappedAPI) KeepAlive() error {
	args := m.Called()

	return args.Error(0)
}

func (m *MockWrappedAPI) Register(regcode string) error {
	args := m.Called(regcode)

	return args.Error(0)
}
func (m *MockWrappedAPI) RegisterOrKeepAlive(regcode string) error {
	args := m.Called(regcode)

	return args.Error(0)
}

// If you need to access the underlying connection object you are
// in API territory. Check MockConnectionWithCredentials to see how
// to mock the connection object. Use
//
//	conn := MockConnectionWithCredentials()
//	api.On("GetConnection").Return(conn)
//
// to get access to the connection object directly.
func (m *MockWrappedAPI) GetConnection() connection.Connection {
	args := m.Called()

	return args.Get(0).(connection.Connection)
}
