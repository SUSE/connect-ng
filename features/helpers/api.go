package helpers

import (
	"testing"

	"github.com/SUSE/connect-ng/internal/credentials"
	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/SUSE/connect-ng/pkg/labels"
	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/stretchr/testify/assert"
)

type ValidationAPI struct {
	t    *testing.T
	conn connection.Connection
}

func NewValidationAPI(t *testing.T) *ValidationAPI {
	return &ValidationAPI{
		t: t,
	}
}

func (api *ValidationAPI) ensureConnection() {
	if api.conn == nil {
		assert.FailNow(api.t, "Try to establish connection to the API before I'm ready!")
	}
}

func (api *ValidationAPI) FetchCredentials() {
	creds, readErr := credentials.ReadCredentials(credentials.SystemCredentialsPath("/"))

	if readErr != nil {
		assert.FailNow(api.t, "Can not read credentials while I should be able: %s", readErr)
		return
	}

	conn := connection.New(connection.DefaultOptions("connect-integration-tests", "0.0.0", "us"), creds)
	api.conn = conn

}

func (api *ValidationAPI) CurrentLabels() []string {
	fetched := []string{}
	api.ensureConnection()

	assignedLabels, readErr := labels.ListLabels(api.conn)

	if readErr != nil {
		assert.FailNow(api.t, "Tried to fetch labels from the API but failed", readErr)
	}

	for _, label := range assignedLabels {
		fetched = append(fetched, label.Name)
	}

	return fetched
}

func (api *ValidationAPI) ProductTree(identifier, version, arch string) *registration.Product {
	api.ensureConnection()

	tree, readErr := registration.FetchProductInfo(api.conn, identifier, version, arch)

	if readErr != nil {
		assert.FailNow(api.t, "Tried to fetch product tree from the API but failed", readErr)
	}

	return tree
}

func (api *ValidationAPI) Activations() []*registration.Activation {
	api.ensureConnection()

	activations, readErr := registration.FetchActivations(api.conn)

	if readErr != nil {
		assert.FailNow(api.t, "Tried to fetch activations from the API but failed", readErr)
	}

	return activations
}
