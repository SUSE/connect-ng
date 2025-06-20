package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/SUSE/connect-ng/pkg/registration"
)

func main() {
	fmt.Println("validate-offline-certificate: Validate a offline registration certificate and print useful information")

	if len(os.Args) != 4 {
		fmt.Println("./validate-offline-certificate <filename> <regcode> <uuid>")
		os.Exit(1)
	}

	path := os.Args[1]
	regcode := os.Args[2]
	uuid := os.Args[3]

	hdl, openErr := os.Open(path)

	if openErr != nil {
		fmt.Printf("Reading %s failed: %s\n", path, openErr)
		os.Exit(1)
	}

	reader := bufio.NewReader(hdl)
	cert, readErr := registration.OfflineCertificateFrom(reader, true)

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

	regcodeMatches, err := cert.RegcodeMatches(regcode)
	if err != nil {
		fmt.Printf("err: %s", err)
		os.Exit(1)
	}
	fmt.Printf("regcode matches: %v\n", regcodeMatches)

	uuidMatches, err := cert.UUIDMatches(uuid)
	if err != nil {
		fmt.Printf("err: %s", err)
		os.Exit(1)
	}
	fmt.Printf("uuid matches: %v\n", uuidMatches)

	fmt.Printf("subscription: %s\n", payload.SubscriptionInfo.Name)
	fmt.Printf("expires: %s\n", payload.SubscriptionInfo.ExpiresAt)
}
