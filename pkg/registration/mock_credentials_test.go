package registration_test

import "github.com/stretchr/testify/mock"

type mockCredentials struct {
	mock.Mock
}

func (m *mockCredentials) HasAuthentication() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockCredentials) Login() (string, string, error) {
	args := m.Called()
	return args.String(0), args.String(1), args.Error(2)
}

func (m *mockCredentials) SetLogin(login, password string) error {
	args := m.Called(login, password)
	return args.Error(0)
}

func (m *mockCredentials) Token() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *mockCredentials) UpdateToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}
