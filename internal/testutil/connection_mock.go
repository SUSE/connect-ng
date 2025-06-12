package testutil

import (
	"io"
	"net/http"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/stretchr/testify/mock"
)

type MockConnection struct {
	mock.Mock
	real connection.Connection
}

func MockConnectionWithCredentials() (*MockConnection, *MockCredentials) {
	creds := &MockCredentials{}
	conn := NewMockConnection(creds, "testing")

	creds.On("Token").Return("sample-token", nil)
	creds.On("Login").Return("sample-login", "sample-password", nil)

	creds.On("UpdateToken", mock.Anything).Return(nil)

	conn.On("GetCredentials").Return(creds)

	return conn, creds
}

func NewMockConnection(creds connection.Credentials, hostname string) *MockConnection {
	opts := connection.DefaultOptions("testing", "---", "---")
	opts.URL = "http://local-testing/"

	conn := connection.New(opts, creds)

	return &MockConnection{
		real: conn,
	}
}

func (m *MockConnection) BuildRequest(verb, path string, body any) (*http.Request, error) {
	request, err := m.real.BuildRequest(verb, path, body)
	return request, err
}

func (m *MockConnection) BuildRequestRaw(verb, path string, body io.Reader) (*http.Request, error) {
	request, err := m.real.BuildRequestRaw(verb, path, body)
	return request, err
}

func (m *MockConnection) Do(request *http.Request) ([]byte, error) {
	args := m.Called(request)

	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockConnection) GetCredentials() connection.Credentials {
	args := m.Called()

	return args.Get(0).(connection.Credentials)
}
