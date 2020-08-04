// +build !go1.13

package httpsutil

import "net/http"

func cloneDefaultTransport() *http.Transport {
	// http.DefaultTransport might be unsafe for a direct-copy since it
	// contains unexported fields: create a new instance using the same
	// initial values.
	//
	// Use a safe noop call to force transport to be initialized. This way
	// the transport has a non-nil TLSClientConfig that we can modify below
	// and retains transparent HTTP/2 support without us needing to
	// configure it explicitly.
	//
	// Only works with Go versions older than 1.13.
	defaultTransport := http.DefaultTransport.(*http.Transport)
	transport := &http.Transport{
		Proxy:                 defaultTransport.Proxy,
		DialContext:           defaultTransport.DialContext,
		MaxIdleConns:          defaultTransport.MaxIdleConns,
		IdleConnTimeout:       defaultTransport.IdleConnTimeout,
		TLSHandshakeTimeout:   defaultTransport.TLSHandshakeTimeout,
		ExpectContinueTimeout: defaultTransport.ExpectContinueTimeout,
	}
	transport.CloseIdleConnections()
	return transport
}
