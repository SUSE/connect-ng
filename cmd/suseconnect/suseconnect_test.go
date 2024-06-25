package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestProcessTokenWithFileToken(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

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
	tempFile, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	fileToken := fmt.Sprintf("@%s", tempFile.Name())
	result, err := processToken(fileToken)
	assert.Error(t, err)
	assert.Equal(t, "", result)
}
func TestProcessTokenWithStdinError(t *testing.T) {
	stdinBackup := os.Stdin
	defer func() { os.Stdin = stdinBackup }()
	r, w, _ := os.Pipe()
	r.Close()

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
	r, w, _ := os.Pipe()
	os.Stdin = r

	// Write to the write end of the pipe in a separate goroutine
	go func() {
		defer w.Close()
		w.Write([]byte("testToken\n"))
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
	tempFile, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(" testToken ")
	if err != nil {
		t.Fatal(err)
	}

	fileToken := fmt.Sprintf("@%s", tempFile.Name())
	result, err := processToken(fileToken)
	assert.NoError(t, err)
	assert.Equal(t, " testToken ", result)
}
