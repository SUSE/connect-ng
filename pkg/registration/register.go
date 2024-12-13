package registration

import (
	"fmt"

	"github.com/SUSE/connect-ng/internal/collectors"
	"github.com/SUSE/connect-ng/pkg/connection"
)

// Register a system by using the given regcode. You also need to provide the
// hostname of the system plus extra info that is to be bundled when registering
// a system.
func Register(conn connection.Connection, regcode, hostname string, info collectors.Result) error {
	var payload struct {
		Hostname string            `json:"hostname"`
		Hwinfo   collectors.Result `json:"hwinfo"`
	}
	payload.Hostname = hostname
	payload.Hwinfo = info

	// TODO: convert to json and pass it to GetRequest

	req, err := conn.GetRequest("POST", "/connect/subscriptions/systems", []byte{})
	if err != nil {
		return err
	}

	body, err := conn.Do(req)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", body)

	return nil
}

// De-register the system pointed by the given authorized connection.
func Deregister(conn connection.Connection) error {
	return nil
}
