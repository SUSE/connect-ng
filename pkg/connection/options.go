package connection

import (
	"crypto/x509"
	"net/http"
	"net/url"
	"time"
)

// Function to provide custom proxy configuration on API connections (e.g. read proxy
// credentials from environment variables, or read proxy credentials from a $HOME/.curlrc file).
type ProxyCallbackFunc func(*http.Request) (*url.URL, error)

const (
	// Default URL to be used for Options, which simply points to the reference
	// SCC server.
	DefaultBaseURL = "https://scc.suse.com"
	DefaultTimeout = 60 * time.Second
)

// Options that are needed in order to form connections to the SCC API. See the
// `ApiConnection` struct.
type Options struct {
	// URL to be used for connections.
	URL string

	// Optional callback that enables connections to establish configuration for
	// HTTP proxies.
	Proxy ProxyCallbackFunc

	// Set a certificate to allow connections to TLS enabled API hosts with self
	// signed certificates.
	Certificate *x509.Certificate

	// True if a secure connection is required.
	Secure bool

	// Name of the application utilizing the library. This will added to the user
	// agent string.
	AppName string

	// Version of the application utilizing the library. Will be added to the user
	// agent string.
	Version string

	// Language used to display API error messages. Empty string means no specific
	// language settings.
	PreferedLanguage string

	// Timeout on how long to wait for an API response
	Timeout time.Duration
}

// Returns the Options suitable for targeting the SCC reference server.
func DefaultOptions(appName, version, language string) Options {
	return Options{
		URL:              DefaultBaseURL,
		Secure:           true,
		AppName:          appName,
		Version:          version,
		PreferedLanguage: language,
		Timeout:          DefaultTimeout,
	}
}
