package connect

import (
	"os"
	"testing"
)

func createTestUptimeLogFileWithContent(content string) (string, error) {
	tempFile, err := os.CreateTemp("", "testUptimeLog")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()
	tempFilePath := tempFile.Name()
	if _, err := tempFile.WriteString(content); err != nil {
		os.Remove(tempFilePath)
		return "", err
	}

	return tempFilePath, nil
}

func TestUptimeLogFileDoesNotExist(t *testing.T) {
	tempFilePath, err := createTestUptimeLogFileWithContent("")
	if err != nil {
		t.Fatalf("Failed to create temp uptime log file for testing")
	}
	os.Remove(tempFilePath)
	uptimeLog, err := readUptimeLogFile(tempFilePath)
	if uptimeLog != nil && err != nil {
		t.Fatalf("Expected uptime log and err to be nil if uptime log file does not exist")
	}
}

func TestReadUptimeLogFile(t *testing.T) {
	uptimeLogFileContent := `2024-01-18:000000000000001000110000
2024-01-13:000000000000000000010000`
	tempFilePath, err := createTestUptimeLogFileWithContent(uptimeLogFileContent)
	if err != nil {
		t.Fatalf("Failed to create temp uptime log file for testing")
	}
	defer os.Remove(tempFilePath)
	uptimeLog, err := readUptimeLogFile(tempFilePath)
	if err != nil {
		t.Fatalf("Failed to read uptime log file: %s", err)
	}
	if uptimeLog == nil {
		t.Fatal("Failed to open uptime log file")
	}
	if len(uptimeLog) != 2 {
		t.Fatalf("Expected two entries in uptime log, got %#v instead", len(uptimeLog))
	}
}
