package main

import (
	"os"
	"strings"
	"testing"
)

func TestEmptyToken(t *testing.T) {
	token, err := processToken("")

	if err != nil {
		t.Fatalf("expecting no errors, but '%v' was given", err)
	}
	if token != "" {
		t.Fatalf("expecting an empty token, but '%v' was given", token)
	}
}

func TestProcessTokenRaw(t *testing.T) {
	token, err := processToken("token")

	if err != nil {
		t.Fatalf("expecting no errors, but '%v' was given", err)
	}
	if token != "token" {
		t.Fatalf("expecting 'token', but '%v' was given", token)
	}
}

func TestProcessTokenFromStdin(t *testing.T) {
	// Setup our mock stdin.
	tempFile, err := os.CreateTemp("", "suseconnect-test-stdin")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	expectedToken := "stdinToken\n"
	if _, err := tempFile.WriteString(expectedToken); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Save the stdin from the OS and restore it before leaving.
	file, err := os.Open(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	oldStdin := os.Stdin
	defer func() {
		file.Close()
		os.Stdin = oldStdin
	}()
	os.Stdin = file

	// Now the actual test :)
	result, err := processToken("-")
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}
	expectedToken = strings.TrimSpace(expectedToken)
	if result != expectedToken {
		t.Errorf("Expected token to be '%s', but got '%s'", expectedToken, result)
	}
}

func TestProcessTokenFromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "tokenfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	expectedToken := "fileToken"
	if _, err := tmpFile.WriteString(expectedToken + "\n"); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	token := "@" + tmpFile.Name()
	result, err := processToken(token)
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}
	if result != expectedToken {
		t.Errorf("Expected token to be '%s', but got '%s'", expectedToken, result)
	}
}

func TestProcessTokenFromBadFile(t *testing.T) {
	token := "@/lala/error/whatever"
	prefix := "failed to open token file"

	if _, err := processToken(token); err == nil {
		t.Fatalf("Expecting an error")
	} else if !strings.HasPrefix(err.Error(), prefix) {
		t.Fatalf("Bad error message; got '%v'", err)
	}
}
