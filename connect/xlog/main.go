package xlog

import (
	"log"
	"os"
)

var (
	Debug *log.Logger
    Error *log.Logger
)

func init() {
	devNullFile, _ := os.Open(os.DevNull)
	Debug = log.New(devNullFile, "", 0)
    Error = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func EnableDebug() {
	Debug = log.New(os.Stderr, "", 0)
}
