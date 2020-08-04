// +build !debug

package https

import "net/http"

// handleDebug handles /debug/ requests if built using a debug build tag.
func handleDebug(handler http.Handler) http.Handler {
	return handler // noop
}
