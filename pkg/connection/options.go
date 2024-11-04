package connection

import (
	"net/http"
	"net/url"
)

// Function to be provided by implementers that want to have some proxy
// configuration on API connections (e.g. read proxy credentials from
// environment variables, or read proxy credentials from a $HOME/.curlrc file).
type ProxyCallbackFunc func(*http.Request) (*url.URL, error)

const (
	// Default URL to be used for Options, which simply points to the reference
	// SCC server.
	DefaultBaseURL = "https://scc.suse.com"
)

// Options that are needed in order to form connections to the SCC API. See the
// `ApiConnection` struct.
type Options struct {
	// URL to be used for connections.
	Url string

	// Optional callback that enables connections to establish configuration for
	// HTTP proxies.
	Proxy ProxyCallbackFunc

	// True if a secure connection is required.
	Secure bool
}

// Returns the Options suitable for targeting the SCC reference server.
func SCCOptions() Options {
	return Options{
		Url:    DefaultBaseURL,
		Secure: true,
	}
}
