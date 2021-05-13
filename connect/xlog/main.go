package xlog

import (
	"log"
	"os"
)

var (
	Debug *log.Logger
)

func init() {
	devNullFile, _ := os.Open(os.DevNull)
	Debug = log.New(devNullFile, "", 0)
}

func EnableDebug() {
	Debug = log.New(os.Stderr, "", 0)
}
