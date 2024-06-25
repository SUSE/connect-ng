package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestProcessTokenWithFileToken(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := os.Remove(tempFile.Name())
		if err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()
	_, err = tempFile.WriteString("testToken")
	if err != nil {
		t.Fatal(err)
	}

	fileToken := fmt.Sprintf("@%s", tempFile.Name())
	result, err := processToken(fileToken)
	assert.NoError(t, err)
	assert.Equal(t, "testToken", result)
}
func TestProcessTokenWithNormalToken(t *testing.T) {
	result, err := processToken("testToken")
	assert.NoError(t, err)
	assert.Equal(t, "testToken", result)
}

func TestProcessTokenWithNonExistentFile(t *testing.T) {
	fileToken := "@non_existent_file"
	result, err := processToken(fileToken)
	assert.Error(t, err)
	assert.Equal(t, "", result)
}

func TestProcessTokenWithEmptyFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := os.Remove(tempFile.Name())
		if err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	fileToken := fmt.Sprintf("@%s", tempFile.Name())
	result, err := processToken(fileToken)
	assert.Error(t, err)
	assert.Equal(t, "", result)
}
func TestProcessTokenWithStdinError(t *testing.T) {
	// Backup and restore stdin
	stdinBackup := os.Stdin
	defer func() { os.Stdin = stdinBackup }()

	// Create a pipe
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Close the read end of the pipe and check for errors
	err = r.Close()
	if err != nil {
		t.Fatalf("Failed to close read end of pipe: %v", err)
	}

	// Set the write end as stdin
	os.Stdin = w

	result, err := processToken("-")
	assert.Error(t, err)
	assert.Equal(t, "", result)
}

func TestProcessTokenWithStdin(t *testing.T) {
	// Backup and restore stdin
	stdinBackup := os.Stdin
	defer func() { os.Stdin = stdinBackup }()

	// Create a pipe and set the read end as stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdin = r

	// Write to the write end of the pipe in a separate goroutine
	go func() {
		defer func() {
			if err := w.Close(); err != nil {
				t.Errorf("Failed to close pipe: %v", err)
			}
		}()
		_, err := w.Write([]byte("testToken\n"))
		if err != nil {
			t.Errorf("Failed to write to pipe: %v", err)
		}
	}()

	result, err := processToken("-")
	assert.NoError(t, err)
	assert.Equal(t, "testToken", result)
}

func TestProcessTokenWithEmptyString(t *testing.T) {
	result, err := processToken("")
	assert.NoError(t, err)
	assert.Equal(t, "", result)
}

// TestProcessTokenWithSpecialCharacters tests the processToken function with a token that contains special characters
func TestProcessTokenWithSpecialCharacters(t *testing.T) {
	result, err := processToken("!@#$%^&*()")
	assert.NoError(t, err)
	assert.Equal(t, "!@#$%^&*()", result)
}

func TestProcessTokenWithFileTokenWhitespace(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := os.Remove(tempFile.Name())
		if err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	_, err = tempFile.WriteString(" testToken ")
	if err != nil {
		t.Fatal(err)
	}

	fileToken := fmt.Sprintf("@%s", tempFile.Name())
	result, err := processToken(fileToken)
	assert.NoError(t, err)
	assert.Equal(t, " testToken ", result)
}
