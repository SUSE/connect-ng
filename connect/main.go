package connect

import (
	"io"
	"log"
	"os"
)

var (
	// CFG is the global struct for config
	CFG   Config
	Debug *log.Logger
	Error *log.Logger
)

func init() {
	Debug = log.New(io.Discard, "", 0)
	Error = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func EnableDebug() {
	Debug.SetOutput(os.Stderr)
}
