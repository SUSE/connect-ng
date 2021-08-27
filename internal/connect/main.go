package connect

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	// AppName holds the name of this SUSE connect client
	AppName = "SUSEConnect-ng" // REVISIT
)

var (
	// CFG is the global struct for config
	CFG Config
	// Debug logger for debugging output
	Debug *log.Logger
)

func init() {
	Debug = log.New(io.Discard, "", 0)
}

// EnableDebug turns on debugging output
func EnableDebug() {
	Debug.SetOutput(os.Stderr)
}

func greenText(text string) string {
	return fmt.Sprintf("\x1b[32m%s\x1b[0m", text)
}

func redText(text string) string {
	return fmt.Sprintf("\x1b[31m%s\x1b[0m", text)
}

func bold(text string) string {
	return fmt.Sprintf("\x1b[1m%s\x1b[0m", text)
}
