package search

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func fixture(t *testing.T, path string) []byte {
	t.Helper()

	absolut, pathErr := filepath.Abs(filepath.Join("../../testdata", path))
	if pathErr != nil {
		t.Fatalf("Could not build fixture path from %s", path)
	}

	data, err := os.ReadFile(absolut)
	if err != nil {
		t.Fatalf("Could not read fixture: %s", err)
	}
	return data

}

func TestPackageSearch(t *testing.T) {
	assert := assert.New(t)

	_, conn, _ := connection.NewMockConnectionWithCredentials()
	response := fixture(t, "pkg/search/search_sles.json")

	conn.On("Do", mock.Anything).Return(response, nil)

	res, err := Package(conn, "SLES", "SLES/15.6/x86_64")
	assert.NoError(err)

	assert.Equal(16, len(res))
	assert.Equal("patterns-lp-lp_sles", res[0].Name)
	assert.Equal(1, len(res[0].Products))
	assert.Equal("SUSE Linux Enterprise Live Patching", res[0].Products[0].Name)
}

func TestUnsupportedServer(t *testing.T) {
	assert := assert.New(t)

	_, conn, _ := connection.NewMockConnectionWithCredentials()

	conn.On("Do", mock.Anything).Return([]byte{}, &connection.ApiError{
		Code:             http.StatusNotFound,
		Message:          "",
		LocalizedMessage: "",
	})

	_, err := Package(conn, "SLES", "SLES/15.6/x86_64")
	assert.True(strings.HasPrefix(err.Error(), "SUSE::Connect::UnsupportedOperation"), true)
}

func TestUnknownError(t *testing.T) {
	assert := assert.New(t)

	_, conn, _ := connection.NewMockConnectionWithCredentials()

	conn.On("Do", mock.Anything).Return([]byte{}, errors.New("whatever"))

	_, err := Package(conn, "SLES", "SLES/15.6/x86_64")
	assert.Equal(err.Error(), "whatever")
}
