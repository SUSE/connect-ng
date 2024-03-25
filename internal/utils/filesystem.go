package utils

import (
    "os"

    "github.com/SUSE/connect-ng/internal/logging"
)

func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func RemoveFile(path string) error {
	logging.Debug.Print("Removing file: ", path)
	if !FileExists(path) {
		return nil
	}
	return os.Remove(path)
}
