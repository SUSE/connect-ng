package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/SUSE/connect-ng/pkg/registration"
)

const (
	hostname = "public-api-demo"
)

func bold(format string, args ...interface{}) {
	fmt.Printf("\033[1m"+format+"\033[0m", args...)
}

func waitForUser(message string) {
	if os.Getenv("NON_INTERACTIVE") != "true" {
		bold("\n%s. Enter to continue\n", message)
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	} else {
		bold("\n%s", message)
	}
}

func runDemo(identifier, version, arch, regcode string) error {
	opts := connection.DefaultOptions("public-api-demo", "1.0", "DE")
	isProxy := false

	if url := os.Getenv("SCC_URL"); url != "" {
		opts.URL = url
		isProxy = true
	}

	bold("1) Setup connection and perform an request\n")
	conn := connection.New(opts, &SccCredentials{})

	// Proxies do not implement /connect/subscriptions/info so we skip it
	if !isProxy {
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
	}

	bold("2) Registering a client to SCC with a registration code\n")
	id, regErr := registration.Register(conn, regcode, hostname, registration.NoSystemInformation, registration.NoExtraData)
	if regErr != nil {
		return regErr
	}
	bold("!! check https://scc.suse.com/systems/%d\n", id)

	bold("3) Activate %s/%s/%s\n", identifier, version, arch)
	_, root, rootErr := registration.Activate(conn, identifier, version, arch, regcode)
	if rootErr != nil {
		return rootErr
	}
	bold("++ %s activated\n", root.FriendlyName)
	waitForUser("Registration complete")

	bold("4) System status // Ping\n")
	systemInformation := registration.SystemInformation{
		"uname": "public api demo - ping",
	}

	extraData := registration.ExtraData{
		"instance_data": "<document>{}</document>",
	}

	status, statusErr := registration.Status(conn, hostname, systemInformation, extraData)
	if statusErr != nil {
		return statusErr
	}

	if status != registration.Registered {
		return errors.New("Could not finalize registration!")
	}
	waitForUser("System update complete")

	bold("5) Activate recommended extensions/modules\n")
	product, prodErr := registration.FetchProductInfo(conn, identifier, version, arch)
	if prodErr != nil {
		return prodErr
	}

	activator := func(ext registration.Product) (bool, error) {
		if ext.Free && ext.Recommended {
			_, act, activateErr := registration.Activate(conn, ext.Identifier, ext.Version, ext.Arch, "")
			if activateErr != nil {
				return false, activateErr
			}
			bold("++ %s activated\n", act.FriendlyName)
			return true, nil
		}
		return false, nil
	}

	if err := product.TraverseExtensions(activator); err != nil {
		return err
	}
	waitForUser("System fully activated")

	bold("6) Show all activations\n")
	activations, actErr := registration.FetchActivations(conn)

	if actErr != nil {
		return actErr
	}

	for i, activation := range activations {
		fmt.Printf("[%d] %s\n", i, activation.Product.Name)
	}
	waitForUser("All activated products are listed")

	bold("7) Deregistration of the client\n")
	if err := registration.Deregister(conn); err != nil {
		return err
	}
	bold("-- System deregistered")
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

	err := runDemo(os.Args[1], os.Args[2], os.Args[3], regcode)

	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}
