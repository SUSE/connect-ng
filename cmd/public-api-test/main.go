package main

import (
	"fmt"
	"os"

	"github.com/SUSE/connect-ng/pkg/connection"
)

const (
	hostname = "public-api-demo"
)

func bold(format string, args ...interface{}) {
	fmt.Printf("\033[1m"+format+"\033[0m", args...)
}

func runDemo(regcode string) error {
	opts := connection.DefaultOptions("public-api-demo", "1.0", "DE")

	if url := os.Getenv("SCC_URL"); url != "" {
		opts.URL = url
	}

	bold("1) Setup connection and perform an request\n")
	conn := connection.New(opts, &SccCredentials{})

	request, buildErr := conn.BuildRequest("GET", "/connect/subscriptions/info", nil)
	if buildErr != nil {
		return buildErr
	}

	connection.AddRegcodeAuth(request, regcode)

	payload, err := conn.Do(request)
	if err != nil {
		return err
	}
	fmt.Printf("!! len(payload): %d characters\n", len(payload))
	fmt.Printf("!! first 40 characters: %s\n", string(payload[0:40]))

	return nil
}

func main() {
	fmt.Println("public-api-demo: A connect client library demo")

	if len(os.Args) != 4 {
		fmt.Println("./public-api-demo IDENTIFIER VERSION ARCH")
		return
	}

	regcode := os.Getenv("REGCODE")
	if regcode == "" {
		fmt.Printf("ERROR: Requireing REGCODE to set as environment variable\n")
		os.Exit(1)
	}

	err := runDemo(regcode)

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}
