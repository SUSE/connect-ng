package main

import (
	"flag"
	"fmt"
	"gitlab.suse.de/doreilly/go-connect/connect"
	"os"
)

var usageHeader = `Usage: SUSEConnect [options]
Register SUSE Linux Enterprise installations with the SUSE Customer Center.
Registration allows access to software repositories (including updates)
and allows online management of subscriptions and organizations.

Manage subscriptions at https://scc.suse.com
`

func main() {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "Root privileges are required to register products and change software repositories.")
		os.Exit(1)
	}

	var status, statusText bool
	flag.BoolVar(&status, "status", false, "Get current system registration status in json format.")
	flag.BoolVar(&status, "s", false, "Get current system registration status in json format.")
	flag.BoolVar(&statusText, "status-text", false, "Get current system registration status in text format.")

	// this function can be changed to display exactly what the ruby SUSEConnect displays.
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usageHeader)
		flag.PrintDefaults()
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
	if statusText {
		fmt.Print(connect.GetProductStatuses("text"))
		return
	}
	// unknown args
	flag.Usage()
	os.Exit(1)
}
