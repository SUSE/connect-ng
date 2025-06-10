package main

import (
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

	if len(os.Args) != 6 {
		fmt.Println("./offline-register-api IDENTIFIER VERSION ARCH REGCODE UUID")
		os.Exit(1)
	}

	identifier := os.Args[1]
	version := os.Args[2]
	arch := os.Args[3]
	regcode := os.Args[4]
	uuid := os.Args[5]

	information := map[string]any{
		"cpus":    8,
		"sockets": 2,
		"vcpus":   16,
	}

	conn := connection.New(opts, connection.NoCredentials{})
	request := registration.BuildOfflineRequest(identifier, version, arch, uuid, information)

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

	fmt.Printf("Certificate validity: %v\n", valid)
	fmt.Printf("Expires at: %v\n", expires)
	fmt.Printf("RegcodeMatches: %v\n", matches)
}
