package connect

import (
	_ "embed"
	"strings"
)

var (
	//go:embed version.txt
	version string
)

// GetShortenedVersion returns the short program version
func GetShortenedVersion() string {
	return strings.Split(version, "~")[0]
}
