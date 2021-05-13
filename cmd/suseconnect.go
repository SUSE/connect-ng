package main

import (
	_ "embed"
	"flag"
	"fmt"
	"gitlab.suse.de/doreilly/go-connect/connect"
	"os"
)

var (
	//go:embed usage.txt
	usageText string
)

func main() {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "Root privileges are required to register products and change software repositories.")
		os.Exit(1)
	}

	var status bool
	flag.BoolVar(&status, "status", false, "")
	flag.BoolVar(&status, "s", false, "")

	// display help like the ruby SUSEConnect
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usageText)
	}
	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}
	flag.Parse()

	if status {
		fmt.Println(connect.GetProductStatuses("json"))
		return
	}
	// unknown args
	flag.Usage()
	os.Exit(1)
}
