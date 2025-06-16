package main

import (
	"bufio"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/SUSE/connect-ng/pkg/labels"
	"github.com/SUSE/connect-ng/pkg/registration"
)

const (
	hostname = "public-api-demo"
)

func bold(format string, args ...any) {
	fmt.Printf("\033[1m"+format+"\033[0m", args...)
}

func waitForUser(message string) {
	if os.Getenv("NON_INTERACTIVE") != "true" {
		bold("\n%s. Enter to continue\n", message)
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	} else {
		bold("\n%s\n", message)
	}
}

func runDemo(identifier, version, arch, regcode string) error {
	opts := connection.DefaultOptions("public-api-demo", "1.0", "DE")
	isProxy := false
	creds := &SccCredentials{}

	if url := os.Getenv("REGISTRATION_HOST_URL"); url != "" {
		opts.URL = url
		isProxy = true
	}

	if credentialTracing := os.Getenv("TRACE_CREDENTIAL_UPDATES"); credentialTracing != "" {
		creds.ShowTraces = true
	}

	if certificatePath := os.Getenv("API_CERT"); certificatePath != "" {
		crt, certReadErr := os.ReadFile(certificatePath)
		if certReadErr != nil {
			return certReadErr
		}

		block, _ := pem.Decode(crt)
		if block == nil {
			return fmt.Errorf("Could not decode the servers certificate")
		}

		cert, parseErr := x509.ParseCertificate(block.Bytes)
		if parseErr != nil {
			return parseErr
		}

		// Set the certificate
		opts.Certificate = cert
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

	bold("7) Label management\n")
	toAssign := []labels.Label{
		labels.Label{Name: "public-library-demo", Description: "Demo label created by the public-api-demo executable"},
		labels.Label{Name: "to-be-removed", Description: "Demo label create by the public-api-demo-executable"},
	}

	fmt.Printf("Assigning labels..\n")
	assigned, assignErr := labels.AssignLabels(conn, toAssign)

	if assignErr != nil {
		return assignErr
	}

	fmt.Printf("Newly assigned labels:\n")
	for _, label := range assigned {
		fmt.Printf(" - %d: %s (%s)\n", label.Id, label.Name, label.Description)
	}

	waitForUser("Now lets unassign a label")

	index := slices.IndexFunc(assigned, func(l labels.Label) bool {
		return l.Name == "to-be-removed"
	})

	if index == -1 {
		return fmt.Errorf("Could not find to-be-removed label for this system! Something went wrong!")
	}

	fmt.Printf("Unassign %s (id: %d)..\n", assigned[index].Name, assigned[index].Id)
	_, unassignErr := labels.UnassignLabel(conn, assigned[index].Id)

	if unassignErr != nil {
		return unassignErr
	}

	fmt.Printf("Fetch updated list of labels..\n")
	updated, listErr := labels.ListLabels(conn)

	if listErr != nil {
		return listErr
	}

	fmt.Printf("Up to date list from SCC:\n")
	for _, label := range updated {
		fmt.Printf(" - %d: %s (%s)\n", label.Id, label.Name, label.Description)
	}

	waitForUser("Labels managed")

	bold("8) Deregistration of the client\n")
	if err := registration.Deregister(conn); err != nil {
		return err
	}
	bold("\n-- System deregistered\n")
	return nil
}

func main() {
	fmt.Println("public-api-demo: A connect client library demo")

	if len(os.Args) != 4 {
		fmt.Println("./public-api-demo IDENTIFIER VERSION ARCH")
		return
	}

	// Allow empty regcodes see registering against RMT without a registration
	// code working
	regcode := os.Getenv("REGCODE")

	err := runDemo(os.Args[1], os.Args[2], os.Args[3], regcode)

	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}
