package main

import (
	"errors"
	"fmt"
	"github.com/SUSE/connect-ng/internal/connect"
	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"strings"
	"testing"
)

var processTokenFunc = processToken

var exitCalled bool
var exit = func(code int) {
	exitCalled = true
}

type MockProcessToken struct {
	mock.Mock
}

func (m *MockProcessToken) ProcessToken(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}

func init() {
	processTokenFunc = processToken
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("forced reader error")
}

func TestReadTokenFromErrorValidToken(t *testing.T) {
	inputToken := "validToken\n"
	reader := strings.NewReader(inputToken)
	token, err := readTokenFromReader(reader)
	if err != nil {
		t.Fatalf("Expected no error but got %v", err)
	}
	if token != "validToken" {
		t.Fatalf("Expected token string to be 'validToken' but got '%s'", token)
	}
}

func TestReadTokenFromReader_MultipleNewlines(t *testing.T) {
	input := "firstToken\nsecondToken\n"
	reader := strings.NewReader(input)

	token, err := readTokenFromReader(reader)
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	expected := "firstToken"
	if token != expected {
		t.Errorf("Expected token to be '%s', but got '%s'", expected, token)
	}
}

func TestReadTokenFromReader_EmptyInput(t *testing.T) {
	reader := strings.NewReader("")

	token, err := readTokenFromReader(reader)
	if err == nil {
		t.Fatalf("Expected error, but got none")
	}

	expectedError := "error: token cannot be empty after reading"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', but got '%s'", expectedError, err.Error())
	}

	if token != "" {
		t.Errorf("Expected empty token, but got '%s'", token)
	}
}

func TestReadTokenFromReader_OnlyNewline(t *testing.T) {
	reader := strings.NewReader("\n")

	token, err := readTokenFromReader(reader)
	if err == nil {
		t.Fatalf("Expected error, but got none")
	}

	expectedError := "error: token cannot be empty after reading"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', but got '%s'", expectedError, err.Error())
	}

	if token != "" {
		t.Errorf("Expected empty token, but got '%s'", token)
	}
}

func TestReadTokenFromReader_ErrorProducingReader(t *testing.T) {
	reader := &errorReader{}

	token, err := readTokenFromReader(reader)
	if err == nil {
		t.Fatalf("Expected error, but got none")
	}

	expectedError := "failed to read token from reader: forced reader error"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', but got '%s'", expectedError, err.Error())
	}

	if token != "" {
		t.Errorf("Expected empty token, but got '%s'", token)
	}
}

func TestProcessToken_RegularToken(t *testing.T) {
	token := "myRegularToken"
	result, err := processToken(token)
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	if result != token {
		t.Errorf("Expected token to be '%s', but got '%s'", token, result)
	}
}

func TestProcessToken_TokenFromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "tokenfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	expectedToken := "fileToken\n"
	if _, err := tmpFile.WriteString(expectedToken); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	token := "@" + tmpFile.Name()
	result, err := processToken(token)
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	expectedToken = strings.TrimSpace(expectedToken)
	if result != expectedToken {
		t.Errorf("Expected token to be '%s', but got '%s'", expectedToken, result)
	}
}

func TestProcessToken_NonExistentFile(t *testing.T) {
	token := "@/non/existent/file"
	_, err := processToken(token)
	if err == nil {
		t.Fatalf("Expected error for non-existent file, but got none")
	}

	expectedError := "failed to open token file '/non/existent/file'"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', but got '%v'", expectedError, err)
	}
}

func TestProcessToken_TokenFromStdin(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_stdin")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	expectedToken := "stdinToken\n"
	if _, err := tempFile.WriteString(expectedToken); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	tempFile.Close()

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	file, err := os.Open(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	os.Stdin = file

	result, err := processToken("-")
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	expectedToken = strings.TrimSpace(expectedToken)
	if result != expectedToken {
		t.Errorf("Expected token to be '%s', but got '%s'", expectedToken, result)
	}
}

func TestProcessToken_ErrorReadingStdin(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_stdin_empty")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	tempFile.Close()
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	file, err := os.Open(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	os.Stdin = file

	_, err = processToken("-")
	if err == nil {
		t.Fatalf("Expected error reading from stdin, but got none")
	}

	expectedError := "error: token cannot be empty after reading"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', but got '%v'", expectedError, err)
	}
}

func parseRegistrationTokenWithInjection(token string) {
	if token != "" {
		connect.CFG.Token = token
		processedToken, processTokenErr := processTokenFunc(token)
		if processTokenErr != nil {
			util.Debug.Printf("Error Processing token %+v", processTokenErr)
			exit(1)
		}
		connect.CFG.Token = processedToken
	}
}

func TestParseToken_Success(t *testing.T) {
	mockProcessToken := new(MockProcessToken)
	mockProcessToken.On("ProcessToken", "valid-token").Return("processed-token", nil)

	processTokenFunc = mockProcessToken.ProcessToken

	exitCalled = false

	parseRegistrationTokenWithInjection("valid-token")

	assert.Equal(t, "processed-token", connect.CFG.Token, "Token should be processed correctly")
	assert.False(t, exitCalled, "os.Exit (simulated) should not be called in a successful case")

	mockProcessToken.AssertExpectations(t)
}

func TestParseToken_ProcessTokenError(t *testing.T) {
	mockProcessToken := new(MockProcessToken)
	mockProcessToken.On("ProcessToken", "invalid-token").Return("", errors.New("failed to process token"))

	processTokenFunc = mockProcessToken.ProcessToken

	exitCalled = false

	parseRegistrationTokenWithInjection("invalid-token")

	assert.True(t, exitCalled, "os.Exit (simulated) should be called when processToken fails")
	assert.Equal(t, "", connect.CFG.Token, "Token should not be updated when processToken fails")

	mockProcessToken.AssertExpectations(t)
}

func TestParseToken_EmptyToken(t *testing.T) {
	exitCalled = false
	parseRegistrationTokenWithInjection("")
	assert.Empty(t, connect.CFG.Token, "Token should not be updated when input token is empty")
	assert.False(t, exitCalled, "os.Exit (simulated) should not be called for empty token")
}
