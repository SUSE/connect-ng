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

// Mock for processToken function in the test file.
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
	reader := strings.NewReader("\n") // Input contains only a newline

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
	reader := &errorReader{} // Custom reader that always returns an error

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
func TestReadTokenFromReader_OnlyWhitespace(t *testing.T) {
	reader := strings.NewReader("   \n") // Input contains only spaces and newline

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
	defer os.Remove(tmpFile.Name()) // Clean up the temp file

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
	// Create a temporary file to simulate stdin
	tempFile, err := os.CreateTemp("", "test_stdin")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up the temp file

	// Write the test token to the temporary file
	expectedToken := "stdinToken\n"
	if _, err := tempFile.WriteString(expectedToken); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Close the file so we can reopen it as stdin
	tempFile.Close()

	// Temporarily replace os.Stdin with the temp file
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original stdin
	file, err := os.Open(tempFile.Name())  // Reopen the temp file for reading
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	os.Stdin = file // Set os.Stdin to the temp file

	// Call processToken with "-"
	result, err := processToken("-")
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	// Verify the result
	expectedToken = strings.TrimSpace(expectedToken) // Trim newline
	if result != expectedToken {
		t.Errorf("Expected token to be '%s', but got '%s'", expectedToken, result)
	}
}

func TestProcessToken_ErrorReadingStdin(t *testing.T) {
	// Create a temporary file to simulate stdin
	tempFile, err := os.CreateTemp("", "test_stdin_empty")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up the temp file

	// Close the temp file to simulate EOF
	tempFile.Close()

	// Temporarily replace os.Stdin with the temp file
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original stdin
	file, err := os.Open(tempFile.Name())  // Open the temp file for reading
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	os.Stdin = file // Set os.Stdin to the temp file

	// Call processToken with "-"
	_, err = processToken("-")
	if err == nil {
		t.Fatalf("Expected error reading from stdin, but got none")
	}

	// Check for the specific error about an empty token
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
			exit(1) // Use the mockable exit function
		}
		connect.CFG.Token = processedToken
	}
}

func TestParseToken_Success(t *testing.T) {
	// Mocking processToken using testify's mock package
	mockProcessToken := new(MockProcessToken)
	mockProcessToken.On("ProcessToken", "valid-token").Return("processed-token", nil)

	// Override processTokenFunc to use the mocked function
	processTokenFunc = mockProcessToken.ProcessToken

	// Reset exitCalled to false for the new test
	exitCalled = false

	// Call the function with a valid token
	parseRegistrationTokenWithInjection("valid-token")

	// Assertions
	assert.Equal(t, "processed-token", connect.CFG.Token, "Token should be processed correctly")
	assert.False(t, exitCalled, "os.Exit (simulated) should not be called in a successful case")

	// Verify the mock expectations
	mockProcessToken.AssertExpectations(t)
}

func TestParseToken_ProcessTokenError(t *testing.T) {
	// Mocking processToken to return an error
	mockProcessToken := new(MockProcessToken)
	mockProcessToken.On("ProcessToken", "invalid-token").Return("", errors.New("failed to process token"))

	// Override processTokenFunc to use the mocked function
	processTokenFunc = mockProcessToken.ProcessToken

	// Reset exitCalled to false for the new test
	exitCalled = false

	// Call the function with an invalid token
	parseRegistrationTokenWithInjection("invalid-token")

	// Assertions
	assert.True(t, exitCalled, "os.Exit (simulated) should be called when processToken fails")
	assert.Equal(t, "", connect.CFG.Token, "Token should not be updated when processToken fails")

	// Verify the mock expectations
	mockProcessToken.AssertExpectations(t)
}

func TestParseToken_EmptyToken(t *testing.T) {
	// Reset exitCalled to false for the new test
	exitCalled = false

	// Call the function with an empty token
	parseRegistrationTokenWithInjection("")

	// Assertions
	assert.Empty(t, connect.CFG.Token, "Token should not be updated when input token is empty")
	assert.False(t, exitCalled, "os.Exit (simulated) should not be called for empty token")
}