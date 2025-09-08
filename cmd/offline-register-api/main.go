package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/SUSE/connect-ng/pkg/registration"
)

func main() {
	fmt.Printf("offline-register-api: Creates a offline request and then uploads it to the offline registration API and retrieves an offline registration certificate\n\n")

	opts := connection.DefaultOptions("offline-register-api", "1.0", "us")

	if url := os.Getenv("SCC_URL"); url != "" {
		opts.URL = url
	}

	if len(os.Args) < 4 || len(os.Args) > 6 {
		fmt.Println("./offline-register-api IDENTIFIER VERSION ARCH REGCODE <SYSTEMINFORMATION>")
		os.Exit(1)
	}

	identifier := os.Args[1]
	version := os.Args[2]
	arch := os.Args[3]
	regcode := os.Args[4]

	information := registration.SystemInformation{}

	if len(os.Args) == 6 {
		file := os.Args[5]

		data, readErr := os.ReadFile(file)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "Error reading system information file: %v\n", readErr)
			os.Exit(1)
		}

		parseErr := json.Unmarshal(data, &information)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling system information JSON: %v\n", parseErr)
			os.Exit(1)
		}
	}

	conn := connection.New(opts, connection.NoCredentials{})
	request := registration.BuildOfflineRequest(identifier, version, arch, information)

	blob, encodeErr := request.Base64Encoded()
	if encodeErr != nil {
		fmt.Printf("Could not encode request: %v\n", encodeErr)
		os.Exit(1)
	}

	encoded, readErr := io.ReadAll(blob)
	if readErr != nil {
		fmt.Printf("Could not read request: %v\n", readErr)
		os.Exit(1)
	}

	fmt.Printf("request blob:\n===\n%s\n===\n", encoded)

	certificate, err := registration.RegisterWithOfflineRequest(conn, regcode, request)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	valid, certErr := certificate.IsValid()
	expires, expireErr := certificate.ExpiresAt()
	matches, regcodeErr := certificate.RegcodeMatches(regcode)

	if certErr != nil {
		fmt.Printf("Certificate error: %v\n", certErr)
		os.Exit(1)
	}

	if expireErr != nil {
		fmt.Printf("Expire error: %v\n", expireErr)
		os.Exit(1)
	}

	if regcodeErr != nil {
		fmt.Printf("Matching error: %v\n", certErr)
		os.Exit(1)
	}

	fmt.Printf("certificate blob:\n===\n%s\n===\n", certificate.EncodedPayload)
	fmt.Printf("Payload: %#v\n", certificate.OfflinePayload)
	fmt.Printf("Certificate validity: %v\n", valid)
	fmt.Printf("Expires at: %v\n", expires)
	fmt.Printf("RegcodeMatches: %v\n", matches)
}
