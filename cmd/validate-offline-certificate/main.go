package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/SUSE/connect-ng/pkg/registration"
)

func main() {
	fmt.Println("validate-offline-certificate: Validate a offline registration certificate and print useful information")

	if len(os.Args) != 2 {
		fmt.Println("./validate-offline-certificate <filename>")
		os.Exit(1)
	}

	path := os.Args[1]

	hdl, openErr := os.Open(path)

	if openErr != nil {
		fmt.Printf("Reading %s failed: %s\n", path, openErr)
		os.Exit(1)
	}

	reader := bufio.NewReader(hdl)
	cert, readErr := registration.OfflineCertificateFrom(reader)

	if readErr != nil {
		fmt.Printf("validation error: %s\n", readErr)
		os.Exit(1)
	}

	valid, validationErr := cert.IsValid()

	if validationErr != nil {
		fmt.Printf("validation error: %s\n", validationErr)
		os.Exit(1)
	}

	if valid {
		fmt.Println("valid: yes")
	} else {
		fmt.Println("valid: no")
	}

	payload, extractErr := cert.ExtractPayload()

	if extractErr != nil {
		fmt.Printf("extract error: %s\n", extractErr)
		os.Exit(1)
	}

	fmt.Printf("uuid: %s\n", payload.Information["uuid"])
	fmt.Printf("subscription: %s\n", payload.Subscription.Name)
	fmt.Printf("expires: %s\n", payload.Subscription.ExpiresAt)
}
