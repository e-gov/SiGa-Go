package https

import "net/http"

// statusResponseWriter wraps a net/http.ResponseWriter and stores the status
// code passed to WriteHeader.
//
// Note that since statusResponseWriter embeds only the ResponseWriter
// interface, any optional methods (e.g., net/http.Flusher, net/http.Pusher)
// become unusable. If these are necessary in the future, then
// statusResponseWriter can be extended to detect and pass-through these.
type statusResponseWriter struct {
	http.ResponseWriter
	statusSet  bool
	statusCode int
}

func (s *statusResponseWriter) WriteHeader(statusCode int) {
	s.statusSet = true
	s.statusCode = statusCode
	s.ResponseWriter.WriteHeader(statusCode)
}

func (s *statusResponseWriter) status() int {
	if s.statusSet {
		return s.statusCode
	}
	// ResponseWriter must send 200 OK if WriteHeader was not called.
	return http.StatusOK
}

// cacheControlResponseWriter wraps a net/http.ResponseWriter and sets the
// Cache-Control header to "no-store" if no Cache-Control header already
// exists.
//
// The alternative of simply setting the header before passing ResponseWriter
// to the net/http.Handler fails in the common case where handlers add headers
// instead of setting them (for example net/http/httputil.ReverseProxy does
// this with proxied response headers). This would leave both the old and new
// Cache-Control values in the response.
//
// Instead perform the check and set the header immediately before the response
// headers are written in WriteHeader.
//
// Note that since cacheControlResponseWriter embeds only the ResponseWriter
// interface, any optional methods (e.g., net/http.Flusher, net/http.Pusher)
// become unusable. If these are necessary in the future, then
// cacheControlResponseWriter can be extended to detect and pass-through these.
type cacheControlResponseWriter struct {
	http.ResponseWriter
}

// setCacheControlHeader can be called before WriteHeader to force the header
// to be set earlier. Useful in cases where the cacheControlResponseWriter
// scope ends before WriteHeader is called.
func (c cacheControlResponseWriter) setCacheControlHeader() {
	// setCacheControlHeader does not detect if the Cache-Control header is
	// set as a Trailer instead, in which case both will be sent. However,
	// this is such a corner case that it should be fine.
	const cacheControl = "Cache-Control" // Must be in canonical form.
	if _, ok := c.Header()[cacheControl]; !ok {
		c.Header().Set(cacheControl, "no-store")
	}
}

func (c cacheControlResponseWriter) WriteHeader(statusCode int) {
	c.setCacheControlHeader()
	c.ResponseWriter.WriteHeader(statusCode)
}

func (c cacheControlResponseWriter) Write(p []byte) (int, error) {
	c.setCacheControlHeader() // In case WriteHeader was not explicitly called.
	return c.ResponseWriter.Write(p)
}
