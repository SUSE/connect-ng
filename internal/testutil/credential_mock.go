package testutil

import "github.com/stretchr/testify/mock"

type MockCredentials struct {
	mock.Mock
}

func (m *MockCredentials) HasAuthentication() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCredentials) Login() (string, string, error) {
	args := m.Called()
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockCredentials) SetLogin(login, password string) error {
	args := m.Called(login, password)
	return args.Error(0)
}

func (m *MockCredentials) Token() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockCredentials) UpdateToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}
