package connection

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectionNew(t *testing.T) {
	assert := assert.New(t)

	opts := DefaultOptions("testApp", "1.0", "en_US")
	creds := NoCredentials{}
	conn := New(opts, creds)

	assert.Implements((*Connection)(nil), conn, "implements connection interface")
	assert.Equal(creds, conn.GetCredentials())
}

func TestConnectionBuildRequest(t *testing.T) {
	assert := assert.New(t)

	opts := DefaultOptions("testApp", "1.0", "en_US")
	creds := NoCredentials{}
	conn := New(opts, creds)

	request, err := conn.BuildRequest("GET", "/test/api", nil)

	assert.NoError(err)
	assert.Equal(request.URL.String(), "https://scc.suse.com/test/api")
	assert.Contains(request.Header.Get("User-Agent"), "testApp/1.0")
	assert.Equal(request.Header.Get("Accept"), DefaultAPIVersion)
	assert.Equal(request.Header.Get("Accept-Language"), "en_US")
}

func TestConnectionBuildRequestNoHTMLEscaping(t *testing.T) {
	assert := assert.New(t)

	opts := DefaultOptions("testApp", "1.0", "en_US")
	creds := NoCredentials{}
	conn := New(opts, creds)

	request, err := conn.BuildRequest("GET", "/test/api", "<xml></xml>")
	assert.NoError(err)

	body, _ := io.ReadAll(request.Body)
	// If escaping is not disabled this will convert < and > into \u003c and \u003e
	assert.Equal("\"<xml></xml>\"\n", string(body))
}

func TestConnectionDoGet(t *testing.T) {
	assert := assert.New(t)

	expected := []byte("{ \"test\": \"foobar\" }")

	handler := func(response http.ResponseWriter) {
		response.WriteHeader(http.StatusOK)
		response.Write(expected)
	}
	server := NewTestServerSetupWith(t, "GET", "/test/api", handler)
	defer server.Close()

	opts := DefaultOptions("testApp", "1.0", "en_US")
	opts.URL = server.URL

	creds := NoCredentials{}
	conn := New(opts, creds)

	request, buildErr := conn.BuildRequest("GET", "/test/api", "")
	assert.NoError(buildErr)

	result, doErr := conn.Do(request)
	assert.NoError(doErr)
	assert.Equal(expected, result)
}

func TestConnectionDoError(t *testing.T) {
	assert := assert.New(t)

	handler := func(response http.ResponseWriter) {
		response.WriteHeader(http.StatusUnprocessableEntity)
		response.Write([]byte("{ \"error\": \"error test message\" }"))
	}
	server := NewTestServerSetupWith(t, "GET", "/test/api", handler)
	defer server.Close()

	opts := DefaultOptions("testApp", "1.0", "en_US")
	opts.URL = server.URL

	creds := NoCredentials{}
	conn := New(opts, creds)

	request, buildErr := conn.BuildRequest("GET", "/test/api", "")
	assert.NoError(buildErr)

	_, doErr := conn.Do(request)
	assert.ErrorContains(doErr, "error test message")
}

func TestConnectionDoErrorTranslation(t *testing.T) {
	assert := assert.New(t)

	expected := "Fehler Test Nachricht"

	handler := func(response http.ResponseWriter) {
		response.WriteHeader(http.StatusUnprocessableEntity)
		response.Write([]byte(fmt.Sprintf("{ \"error\": \"error test message\", \"localized_error\": \"%s\" }", expected)))
	}
	server := NewTestServerSetupWith(t, "GET", "/test/api", handler)
	defer server.Close()

	opts := DefaultOptions("testApp", "1.0", "en_US")
	opts.URL = server.URL

	creds := NoCredentials{}
	conn := New(opts, creds)

	request, buildErr := conn.BuildRequest("GET", "/test/api", "")
	assert.NoError(buildErr)

	_, doErr := conn.Do(request)
	assert.ErrorContains(doErr, expected)
}

func TestConnectionUpdateToken(t *testing.T) {
	assert := assert.New(t)

	oldToken := "17daf81b-2eda-42f4-ad6f-86a08ce20341"
	newToken := "8ffe9f9f-18b6-4692-8588-b5aae3c65cc4"

	handler := func(response http.ResponseWriter) {
		response.Header().Add("System-Token", newToken)
		response.WriteHeader(http.StatusOK)
	}
	server := NewTestServerSetupWith(t, "GET", "/test/api", handler)
	defer server.Close()

	opts := DefaultOptions("testApp", "1.0", "en_US")
	opts.URL = server.URL

	creds := &mockCredentials{}
	creds.On("Token").Return(oldToken, nil)
	creds.On("UpdateToken", newToken).Return(nil)

	conn := New(opts, creds)

	request, buildErr := conn.BuildRequest("GET", "/test/api", "")
	assert.NoError(buildErr)

	_, doErr := conn.Do(request)
	assert.NoError(doErr)

	creds.AssertExpectations(t)
}
