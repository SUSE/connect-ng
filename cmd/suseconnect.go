package main

import (
	_ "embed"
	"flag"
	"fmt"
	"gitlab.suse.de/doreilly/go-connect/connect"
	"gitlab.suse.de/doreilly/go-connect/connect/xlog"
	"os"
)

var (
	//go:embed usage.txt
	usageText  string
	status     bool
	statusText bool
	debug      bool
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

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}
	flag.Parse()
	if debug {
		xlog.EnableDebug()
	}
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "Root privileges are required to register products and change software repositories.")
		os.Exit(1)
	}
	xlog.Debug.Println("cmd line:", os.Args)
	if status {
		fmt.Println(connect.GetProductStatuses("json"))
		return
	}
	if statusText {
		fmt.Print(connect.GetProductStatuses("text"))
		return
	}
	// unknown args
	flag.Usage()
	os.Exit(1)
}
