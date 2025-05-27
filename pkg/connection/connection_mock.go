package connection

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockConnection struct {
	mock.Mock
	real Connection
}

func NewMockConnectionWithCredentials() (*MockConnection, *MockCredentials) {
	creds := NewMockCredentials()
	conn := NewMockConnection(creds, "testing")

	creds.On("Token").Return("sample-token", nil)
	creds.On("Login").Return("sample-login", "sample-password", nil)

	creds.On("UpdateToken", mock.Anything).Return(nil)

	conn.On("GetCredentials").Return(creds)

	return conn, creds
}

func NewMockConnection(creds Credentials, hostname string) *MockConnection {
	opts := DefaultOptions("testing", "---", "---")
	opts.URL = "http://local-testing/"

	conn := New(opts, creds)

	return &MockConnection{
		real: conn,
	}
}

func (m *MockConnection) BuildRequest(verb, path string, body any) (*http.Request, error) {

	request, err := m.real.BuildRequest(verb, path, body)
	return request, err
}

func (m *MockConnection) Do(request *http.Request) ([]byte, error) {
	args := m.Called(request)

	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockConnection) GetCredentials() Credentials {
	args := m.Called()

	return args.Get(0).(Credentials)
}
