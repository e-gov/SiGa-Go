package https

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	stdlog "log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/pkg/errors"

	"stash.ria.ee/vis3/vis3-common/pkg/log"
	"stash.ria.ee/vis3/vis3-common/pkg/session"
)

var (
	corsMethods = []string{"DELETE", "GET", "HEAD", "OPTIONS", "POST", "PUT", "PATCH"}
	corsHeaders = []string{"Content-Type", "Disable-Data-Filter"}
)

// key is the key type of values stored in contexts by this package.
type key int

const sessionErrorKey key = 0 // Context key for session header errors.

// Server is a HTTPS server. It is a wrapper around a standard net/http.Server
// instance simplifying configuration, startup, and graceful close.
type Server struct {
	server http.Server
}

// NewServer creates a new HTTPS server from the configuration and handler. It
// wraps the handler with some common filters (e.g., CORS handling, access
// logging, request identifier assignment).
// Should only be used if the session header is not yet verified, i.e. by SEA.
func NewServer(conf ServerConf, handler http.Handler) *Server {
	return newServer(conf, false, handler)
}

// NewServerWithSessionHeader is identical to NewServer, except that it includes
// the filter for setting the SID in the log context.
// Should only be used if the session header can be trusted, i.e. modules behind
// SEA.
func NewServerWithSessionHeader(conf ServerConf, handler http.Handler) *Server {
	return newServer(conf, true, handler)
}

// newServer is the internal implementation of the constructurs.
func newServer(conf ServerConf, hasSession bool, handler http.Handler) *Server {
	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{
			tls.Certificate(conf.TLS),
		},
	}

	// Apply client side TLS authentication, if configured
	if conf.ClientCAs != nil {
		log.Info().Log(nil, "apply_tls_client_auth")
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.ClientCAs = (*x509.CertPool)(conf.ClientCAs)
	}

	return &Server{
		server: http.Server{
			Addr:      conf.Listen,
			Handler:   filter(handler, conf, hasSession),
			TLSConfig: &tlsConfig,

			ReadTimeout:       conf.ReadTimeout.Or(DefaultReadTimeout),
			ReadHeaderTimeout: conf.ReadHeaderTimeout.Or(DefaultReadHeaderTimeout),
			WriteTimeout:      conf.WriteTimeout.Or(DefaultWriteTimeout),
			IdleTimeout:       conf.IdleTimeout.Or(DefaultIdleTimeout),

			ConnState: connState,
			ErrorLog:  stdlog.New(errorLog{}, "", 0),
		},
	}
}

func connState(conn net.Conn, state http.ConnState) {
	var tag string
	switch state {
	case http.StateNew:
		tag = "new"
	case http.StateActive:
		tag = "active"
	case http.StateIdle:
		tag = "idle"
	case http.StateHijacked:
		tag = "hijacked"
	case http.StateClosed:
		tag = "closed"
	}
	log.Debug().WithString("remote", conn.RemoteAddr()).Log(nil, tag)
}

type errorLog struct{}

func (errorLog) Write(p []byte) (n int, err error) {
	log.Error().WithStringf("error", "%s", p).Log(nil, "http_error")
	return len(p), nil
}

// ListenAndServe calls ListenAndServeTLS("", "") on the wrapped server. Refer
// to that documentation for details.
func (s *Server) ListenAndServe() error {
	return errors.WithStack(s.server.ListenAndServeTLS("", ""))
}

// Shutdown gracefully stops the server. If a shutdown is not successful within
// the given timeout, then a forceful stop is performed instead.
func (s *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := errors.Wrap(s.server.Shutdown(ctx), "shutdown server")
	if errors.Cause(err) == context.DeadlineExceeded {
		if cerr := s.server.Close(); cerr != nil {
			err = errors.Wrap(cerr, "close server")
		}
	}
	return err
}

// filter adds common HTTPS server filters to a handler.
func filter(handler http.Handler, conf ServerConf, hasSession bool) http.Handler {
	handler = handleDebug(handler) // Handle /debug/ requests in dev env.

	filters := []func(http.Handler) http.Handler{
		reqIDFilter,
		newSessionFilter(hasSession),
		logFilter,
		sessionErrorFilter,
		newClientsFilter(conf.AllowedClients),
		newCORSFilter(conf.AllowedOrigins),
		cacheControlFilter,
	}
	for i := len(filters) - 1; i >= 0; i-- { // Wrap in reverse order.
		handler = filters[i](handler)
	}
	return handler
}

func skipFilter(next http.Handler) http.Handler { return next }

// reqIDFilter injects a request identifier into the request's context.
func reqIDFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := make([]byte, 32)
		if _, err := rand.Read(reqID); err != nil {
			log.Error().WithError(err).Log(r.Context(), "rand_error")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r = r.WithContext(log.WithRequestID(r.Context(), hex.EncodeToString(reqID)))
		next.ServeHTTP(w, r)
	})
}

// newSessionFilter returns a filter that reads the session header and if
// present, injects the session data and logging SID into the request's
// context.
//
// If reading the session header returns an error, then we want to stop
// processing with StatusBadRequest, but only after logFilter has run. Yet we
// want this filter to be earlier in the chain, so that if we do have a
// session, we can log the identifier in logFilter. To work around this, the
// session error is stored in the request context and sessionErrorFilter will
// check it after logFilter.
//
// If hasSession is false, then newSessionFilter returns skipFilter instead.
func newSessionFilter(hasSession bool) func(http.Handler) http.Handler {
	if !hasSession {
		return skipFilter
	}
	log.Info().Log(nil, "add_filter")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			sess, err := session.GetHeader(r)
			if err != nil {
				ctx = context.WithValue(ctx, sessionErrorKey, err)
			} else if sess.SID != "" {
				ctx = session.WithHeader(ctx, sess)
				ctx = log.WithSessionID(ctx, sess.SID)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// logFilter logs when a request is accepted and its handling is completed.
func logFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := log.Info().
			WithString("remote", r.RemoteAddr).
			WithString("for", r.Header.Get("X-Forwarded-For")).
			WithString("method", r.Method).
			WithString("uri", r.RequestURI).
			WithString("tls.version", r.TLS.Version).
			WithString("tls.cipher", r.TLS.CipherSuite).
			WithString("tls.sni", r.TLS.ServerName)
		if len(r.TLS.PeerCertificates) > 0 {
			msg = msg.WithString("tls.issuer", r.TLS.PeerCertificates[0].Issuer)
			msg = msg.WithString("tls.subject", r.TLS.PeerCertificates[0].Subject)
		}
		msg.Log(r.Context(), "handle")
		srw := statusResponseWriter{ResponseWriter: w}
		next.ServeHTTP(&srw, r)
		log.Info().WithString("status", srw.status()).Log(r.Context(), "done")
	})
}

// sessionErrorFilter complements newSessionFilter by checking for any session
// header errors later in the filter pipeline.
func sessionErrorFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.Context().Value(sessionErrorKey); err != nil {
			log.Error().WithError(err.(error)).Log(r.Context(), "get_session_error")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// newCORSFilter returns a filter that adds common Access-Control headers to
// the response.
//
// If allowedOrigins is empty, then newCORSFilter returns skipFilter instead.
func newCORSFilter(allowedOrigins []string) func(http.Handler) http.Handler {
	if len(allowedOrigins) == 0 {
		return skipFilter
	}
	log.Info().WithString("origins", allowedOrigins).Log(nil, "add_filter")
	return handlers.CORS(
		handlers.AllowedOrigins(allowedOrigins),
		handlers.AllowedHeaders(corsHeaders),
		handlers.AllowedMethods(corsMethods),
		handlers.AllowCredentials(),
	)
}

// cacheControlFilter instructs caches not to store the response by setting the
// Cache-Control header to "no-store" unless the handler has already set some
// other value. The assumption is that we do not wish to cache API responses by
// default.
func cacheControlFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ccrw := cacheControlResponseWriter{ResponseWriter: w}
		next.ServeHTTP(ccrw, r)
		ccrw.setCacheControlHeader() // Ensure header is set even if WriteHeader was not called.
	})
}
