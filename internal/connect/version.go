package connect

import (
	_ "embed"
	"strings"
)

var (
	//go:embed version.txt
	version string
)

func GetShortenedVersion() string {
	return strings.Split(version, "~")[0]
}
