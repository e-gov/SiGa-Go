/*
Package https provides convenience wrappers around the net/http package.
*/
package https

import (
	"time"

	"stash.ria.ee/vis3/vis3-common/pkg/confutil"
)

// Default timeouts used if the corresponding values in Conf are zero.
const (
	DefaultClientTimeout     = 25 * time.Second
	DefaultReadTimeout       = 30 * time.Second
	DefaultReadHeaderTimeout = 15 * time.Second
	DefaultWriteTimeout      = 30 * time.Second
	DefaultIdleTimeout       = 120 * time.Second
)

// ServerConf contains package configuration values. A JSON-encoding of the
// configuration can be directly unmarshaled into an instance of ServerConf.
type ServerConf struct {
	// Listen is the TCP address that the HTTPS server should listen on. It
	// consists of a "host:port" pair.
	//
	// host can either be a hostname, an IPv4 or IPv6 address, or empty.
	// If empty, then the server will listen on all interfaces. Using a
	// hostname is not recommended, because only one of the resolved IP
	// addresses is used.
	//
	// port can either be a service name, a number, or empty. If 0 or
	// empty, then a port number is chosen automatically.
	//
	// If Listen is empty, then ":https" is used.
	Listen string

	// TLS is the certificate chain and private key for the HTTPS server.
	TLS confutil.TLS

	// AllowedOrigins is the optional list of allowed origins for enabling CORS.
	AllowedOrigins []string

	// ClientCAs, if specified, is the set of root certificates for client side TLS verification.
	ClientCAs *confutil.CertPool

	// AllowedClients is the configuration of the TLS client certificate common name filter.
	//
	// The elements of the whitelist in the access-control list are compared to the common
	// name on the client certificate of the incoming HTTP TLS connection.
	//
	// If the whitelist configured for an incoming HTTP request is empty, access is granted.
	// If no matching whitelist is configured for an incoming HTTP request, access is denied.
	AllowedClients AccessControlList

	// ReadTimeout, ReadHeaderTimout, WriteTimeout, and IdleTimeout are
	// server timeout values in seconds. See the godoc for net/http.Server
	// for an exact explanation of each value.
	ReadTimeout       confutil.Seconds `json:"ReadTimeoutSeconds"`
	ReadHeaderTimeout confutil.Seconds `json:"ReadHeaderTimeoutSeconds"`
	WriteTimeout      confutil.Seconds `json:"WriteTimeoutSeconds"`
	IdleTimeout       confutil.Seconds `json:"IdleTimeoutSeconds"`
}

// AccessControlList maps HTTP request path and method pairs to Whitelists.
//
// The structure mimics the endpoint specification of an OpenAPI documentation
// where the first level of keys specifies the HTTP request path (e.g.
// '/api/v1/heartbeat' or '*' to match any path) and the second level the
// method of the HTTP request ('GET', 'POST' etc or '*' to match any method).
//
// The Whitelist contains the subjects that are granted access to the endpoint
// specified by the map keys.
type AccessControlList map[string]map[string]Whitelist

// Whitelist is an abstract list of allowed subjects. The domain of the values can vary by use-case.
// It can be used as a list of Common Names on the TLS client certificate that are whitelisted.
type Whitelist []string

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
