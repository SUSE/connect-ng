package main

import (
	_ "embed"
	"flag"
	"fmt"
	"gitlab.suse.de/doreilly/go-connect/connect"
	"os"
	"strings"
)

var (
	//go:embed usage.txt
	usageText   string
	status      bool
	statusText  bool
	debug       bool
	writeConfig bool
)

func init() {
	// display help like the ruby SUSEConnect
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usageText)
	}

	flag.BoolVar(&status, "status", false, "")
	flag.BoolVar(&status, "s", false, "")
	flag.BoolVar(&statusText, "status-text", false, "")
	flag.BoolVar(&debug, "debug", false, "")
	flag.BoolVar(&writeConfig, "write-config", false, "")
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "Root privileges are required to register products and change software repositories.")
		os.Exit(1)
	}
	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}
	flag.Parse()
	if debug {
		connect.EnableDebug()
	}
	connect.Debug.Println("cmd line:", os.Args)
	connect.Debug.Println("For http debug use: GODEBUG=http2debug=2", strings.Join(os.Args, " "))
	if status {
		fmt.Println(connect.GetProductStatuses("json"))
	} else if statusText {
		fmt.Print(connect.GetProductStatuses("text"))
	}
	if writeConfig {
		if err := connect.CFG.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Problem writing configuration: %s", err)
			os.Exit(1)
		}
	}
}
