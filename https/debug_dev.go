// +build debug

package https

import (
	"net/http"
	"net/http/pprof"
	"strings"
)

// handleDebug handles /debug/ requests if built using a debug build tag.
func handleDebug(next http.Handler) http.Handler {
	debug := http.NewServeMux()

	// Locally replicate the init function of net/http/pprof.
	debug.HandleFunc("/debug/pprof/", pprof.Index)
	debug.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	debug.HandleFunc("/debug/pprof/profile", pprof.Profile)
	debug.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	debug.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/debug/") {
			debug.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
