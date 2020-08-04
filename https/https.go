/*
Package https provides convenience wrappers around the net/http package.
*/
package https

import (
	"time"

	"github.com/e-gov/SiGa-Go/confutil"
)

// Default timeouts used if the corresponding values in Conf are zero.
const (
	DefaultClientTimeout     = 25 * time.Second
	DefaultReadTimeout       = 30 * time.Second
	DefaultReadHeaderTimeout = 15 * time.Second
	DefaultWriteTimeout      = 30 * time.Second
	DefaultIdleTimeout       = 120 * time.Second
)

// ClientConf is the configuration for https.Client
type ClientConf struct {
	// URL is the (base) URL of the HTTP service used by the client.
	URL confutil.URL

	// ClientTLS, if specified, is the certificate chain and private key
	// used for TLS client-authentication when connecting to server.
	ClientTLS confutil.TLS

	// RootCAs, if specified, is the set of root certificates for server side TLS verification.
	RootCAs *confutil.CertPool

	// Timeout is the HTTP timeout of the connections made by the HTTP client.
	Timeout confutil.Seconds `json:"TimeoutSeconds"`
}
