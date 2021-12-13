package connect

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	// Debug logger for debugging output
	Debug *log.Logger
	// Info logger for standard info output
	Info *log.Logger
	// QuietOut is used to simplify --quiet option
	QuietOut *log.Logger
)

func init() {
	Debug = log.New(io.Discard, "", 0)
	Info = log.New(os.Stdout, "", 0)
	QuietOut = log.New(io.Discard, "", 0)
}

// EnableDebug turns on debugging output
func EnableDebug() {
	Debug.SetOutput(os.Stderr)
}

func isLoggerEnabled(l *log.Logger) bool {
	return l.Writer() != io.Discard
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
