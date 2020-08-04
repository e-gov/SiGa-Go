/*
Package httpsutil provides HTTP utility functions in addition to the parent
https and the standard library net/http/httputil packages.
*/
package httpsutil

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"

	"stash.ria.ee/vis3/vis3-common/pkg/confutil"
)

// Transport returns a copy of http.DefaultTransport with custom TLS
// configuration for server and client authentication. The returned Transport
// has transparent HTTP/2 support.
func Transport(rootCAs *confutil.CertPool, clientCert *confutil.TLS) http.RoundTripper {
	if rootCAs == nil && (clientCert == nil || clientCert.Certificate == nil) {
		return http.DefaultTransport
	}

	transport := cloneDefaultTransport()
	transport.TLSClientConfig.RootCAs = (*x509.CertPool)(rootCAs)
	if clientCert != nil && clientCert.Certificate != nil {
		transport.TLSClientConfig.Certificates = []tls.Certificate{
			tls.Certificate(*clientCert),
		}
	}
	return transport
}
