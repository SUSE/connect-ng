package zypper

import (
    "os"
    "strings"
    "fmt"
    "path/filepath"

    "github.com/SUSE/connect-ng/internal/config"
    "github.com/SUSE/connect-ng/internal/utils"
)

const (
	oemPath    = "/var/lib/suseRegister/OEM"
)


// get first line of OEM file if present
func oemReleaseType(productLine string) (string, error) {
	if productLine == "" {
		return "", fmt.Errorf("empty productline")
	}
	oemFile := filepath.Join(config.CFG.FsRoot, oemPath, productLine)
	if !utils.FileExists(oemFile) {
		return "", fmt.Errorf("OEM file not found: %v", oemFile)
	}
	data, err := os.ReadFile(oemFile)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("empty OEM file: %v", oemFile)
	}
	return strings.TrimSpace(lines[0]), nil
}
