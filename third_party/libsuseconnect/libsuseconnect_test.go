package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net"
	"net/url"
	"testing"

	"github.com/SUSE/connect-ng/internal/connect"
	cred "github.com/SUSE/connect-ng/internal/credentials"
	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type timeoutError struct{}

func (timeoutError) Error() string { return "request timed out" }
func (timeoutError) Timeout() bool { return true }

type errorResponse struct {
	ErrType string `json:"err_type"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Data    string `json:"data,omitempty"`
}

func certificatePEM(cert *x509.Certificate) string {
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}))
}

func TestErrorToJSON(t *testing.T) {
	leaf := &x509.Certificate{Raw: []byte("leaf certificate")}
	issuer := &x509.Certificate{Raw: []byte("issuer certificate")}
	invalidCertificateError := x509.CertificateInvalidError{
		Cert:   leaf,
		Reason: x509.Expired,
	}
	unknownAuthorityError := x509.UnknownAuthorityError{Cert: leaf}
	verificationError := &tls.CertificateVerificationError{
		UnverifiedCertificates: []*x509.Certificate{leaf, issuer},
		Err:                    unknownAuthorityError,
	}
	hostnameError := x509.HostnameError{Certificate: leaf, Host: "example.invalid"}
	networkError := &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}
	plainURLError := &url.Error{Op: "Get", URL: "https://example.invalid", Err: errors.New("unexpected response")}
	jsonError := connect.JSONError{Err: errors.New("invalid character")}
	genericError := errors.New("generic failure")

	tests := []struct {
		name string
		err  error
		want errorResponse
	}{
		{
			name: "API error",
			err:  &connection.ApiError{Code: 422, Message: "invalid product"},
			want: errorResponse{ErrType: "APIError", Message: "invalid product", Code: 422},
		},
		{
			name: "timeout",
			err:  &url.Error{Op: "Get", URL: "https://example.invalid", Err: timeoutError{}},
			want: errorResponse{ErrType: "Timeout", Message: "request timed out"},
		},
		{
			name: "invalid certificate",
			err:  &url.Error{Op: "Get", URL: "https://example.invalid", Err: invalidCertificateError},
			want: errorResponse{
				ErrType: "SSLError",
				Message: invalidCertificateError.Error(),
				Code:    10,
				Data:    certificatePEM(leaf),
			},
		},
		{
			name: "unknown certificate authority",
			err:  &url.Error{Op: "Get", URL: "https://example.invalid", Err: unknownAuthorityError},
			want: errorResponse{
				ErrType: "SSLError",
				Message: unknownAuthorityError.Error(),
				Code:    19,
				Data:    certificatePEM(leaf),
			},
		},
		{
			name: "TLS certificate verification",
			err:  &url.Error{Op: "Get", URL: "https://example.invalid", Err: verificationError},
			want: errorResponse{
				ErrType: "SSLError",
				Message: verificationError.Error(),
				Code:    19,
				Data:    certificatePEM(issuer) + certificatePEM(leaf),
			},
		},
		{
			name: "certificate hostname",
			err:  &url.Error{Op: "Get", URL: "https://example.invalid", Err: hostnameError},
			want: errorResponse{
				ErrType: "SSLError",
				Message: hostnameError.Error(),
				Data:    certificatePEM(leaf),
			},
		},
		{
			name: "network operation",
			err:  &url.Error{Op: "Get", URL: "https://example.invalid", Err: networkError},
			want: errorResponse{ErrType: "NetError", Message: networkError.Error()},
		},
		{
			name: "unclassified URL error",
			err:  plainURLError,
			want: errorResponse{Message: plainURLError.Error()},
		},
		{
			name: "JSON decoding",
			err:  jsonError,
			want: errorResponse{ErrType: "JSONError", Message: "invalid character"},
		},
		{
			name: "malformed SCC credentials",
			err:  cred.ErrMalformedSccCredFile,
			want: errorResponse{
				ErrType: "MalformedSccCredentialsFile",
				Message: cred.ErrMalformedSccCredFile.Error(),
			},
		},
		{
			name: "missing credentials file",
			err:  cred.ErrMissingCredentialsFile,
			want: errorResponse{
				ErrType: "MissingCredentialsFile",
				Message: cred.ErrMissingCredentialsFile.Error(),
			},
		},
		{
			name: "generic error",
			err:  genericError,
			want: errorResponse{Message: genericError.Error()},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			var got errorResponse
			err := json.Unmarshal([]byte(errorToJSON(test.err)), &got)
			require.NoError(t, err, "errorToJSON must return valid JSON")

			assert.Equal(test.want, got)
		})
	}
}
