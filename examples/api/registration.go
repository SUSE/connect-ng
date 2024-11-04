package main

import (
	"github.com/SUSE/connect-ng/pkg/registration"
)

func credentialsChanged(reg *Credentials) {
	// TODO: write to disk
}

func credentialsFromFS() *registration.SccCredentials {
	&registration.SccCredentials{}
}

func main() {
	info := collectors.Result
	// NOTE: if you have multiple processes dealing with this library, make sure
	// to enforce a locking mechanism, because SCC's API only works as a
	// single-threaded process.

	cred := credentialsFromFS()

	// Basic stuff.
	conn := registration.NewConnection(cred)
	conn.SetProxy() // Optional

	registration.Announce(conn, "whatever", "hostname", info)

	productsIdentifier := {}
	for _, pi := range productsIdentifier {
		registration.Activate(conn, ProductIdentifier)
	}

	//registration.Deactivate(conn)
	//registration.Deregister(conn)
}
