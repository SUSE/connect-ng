package registration_test

import (
	"net/http"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/stretchr/testify/mock"
)

func mockConnectionWithCredentials() (*mockConnection, *mockCredentials) {
	creds := &mockCredentials{}
	conn := newMockConnection(creds, "testing")

	creds.On("Token").Return("sample-token", nil)
	creds.On("Login").Return("sample-login", "sample-password", nil)

	creds.On("UpdateToken", mock.Anything).Return(nil)

	conn.On("GetCredentials").Return(creds)

	return conn, creds
}

func newMockConnection(creds connection.Credentials, hostname string) *mockConnection {
	opts := connection.DefaultOptions("testing", "---", "---")
	opts.URL = "http://local-testing/"

	conn := connection.New(opts, creds)

	return &mockConnection{
		real: conn,
	}
}

type mockConnection struct {
	mock.Mock
	real connection.Connection
}

func (m *mockConnection) BuildRequest(verb, path string, body any) (*http.Request, error) {

	request, err := m.real.BuildRequest(verb, path, body)
	return request, err
}

func (m *mockConnection) Do(request *http.Request) (int, []byte, error) {
	args := m.Called(request)

	return args.Int(0), args.Get(1).([]byte), args.Error(2)
}

func (m *mockConnection) GetCredentials() connection.Credentials {
	args := m.Called()

	return args.Get(0).(connection.Credentials)
}
