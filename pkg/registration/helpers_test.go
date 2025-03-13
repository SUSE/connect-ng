package registration

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

func matchBody(t *testing.T, body string) func(mock.Arguments) {
	assert := assert.New(t)

	return func(args mock.Arguments) {
		request := args.Get(0).(*http.Request)
		body, readErr := io.ReadAll(request.Body)

		assert.NoError(readErr)
		assert.Equal(strings.TrimSpace(string(body)), string(body), "request.Body matches")
	}
}

func checkAuthByRegcode(t *testing.T, regcode string) func(mock.Arguments) {
	assert := assert.New(t)

	return func(args mock.Arguments) {
		request := args.Get(0).(*http.Request)
		token := request.Header.Get("Authorization")

		expected := fmt.Sprintf("Token token=%s", regcode)
		assert.Equal(expected, token, "regcode is set as authorization header")
	}
}

func checkAuthBySystemCredentials(t *testing.T, login, password string) func(mock.Arguments) {
	assert := assert.New(t)
	encoded := base64.StdEncoding.EncodeToString([]byte(login + ":" + password))

	return func(args mock.Arguments) {
		request := args.Get(0).(*http.Request)
		token := request.Header.Get("Authorization")

		expected := fmt.Sprintf("Basic %s", encoded)
		assert.Equal(expected, token, "system credentials are set as authorization header")
	}
}
