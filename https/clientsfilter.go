package https

import (
	"net/http"
	"strings"

	"stash.ria.ee/vis3/vis3-common/pkg/log"
)

const allowAll = "*"

type clientsFilter struct {
	next      http.Handler
	endpoints map[string]map[string]whitelist
}

// newClientsFilter returns an HTTP filter that applies the TLS client certificate Common Name
// whitelist check before calling the given HTTP handler.
//
// If allowedClients is empty, then newClientsFilter returns skipFilter instead.
func newClientsFilter(allowedClients AccessControlList) func(http.Handler) http.Handler {
	// If the filter is not configured, skip it.
	if len(allowedClients) == 0 {
		return skipFilter
	}

	log.Info().WithJSON("acl", allowedClients).Log(nil, "add_filter")

	endpoints := make(map[string]map[string]whitelist, len(allowedClients))
	for path, methodMap := range allowedClients {
		endpoints[path] = make(map[string]whitelist, len(methodMap))
		for method, wl := range methodMap {
			endpoints[path][strings.ToLower(method)] = wrapWhitelist(wl)
		}
	}
	return func(next http.Handler) http.Handler {
		return clientsFilter{next: next, endpoints: endpoints}
	}
}

func (cf clientsFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wl, ok := cf.whitelist(r)
	if !ok {
		// No whitelist configured for the request: deny it.
		log.Error().Log(r.Context(), "whitelist_not_found_error")
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if !wl.empty() {
		// If whitelist is not empty, check the client TLS certificate.
		// All verified chains start with the same client certificate
		// so only check the name in the first chain.
		var found bool
		if r.TLS != nil && len(r.TLS.VerifiedChains) > 0 {
			chain := r.TLS.VerifiedChains[0]
			if len(chain) > 0 {
				found = wl.contains(chain[0].Subject.CommonName)
			}
		}
		if !found {
			log.Error().
				WithString("whitelist", wl.Whitelist).
				Log(r.Context(), "access_denied_error")
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}
	cf.next.ServeHTTP(w, r)
}

// whitelist returns the whitelist for the specified HTTP request.
func (cf clientsFilter) whitelist(r *http.Request) (whitelist, bool) {
	// 1. Try exact path match.
	if methodMap, ok := cf.endpoints[r.URL.Path]; ok {
		if wl, ok := method(methodMap, r.Method); ok {
			return wl, ok
		}
	}

	// 2. Try path prefix match - required for requests like '/api/v1/resource/{id}'.
	for path, methodMap := range cf.endpoints {
		if strings.HasPrefix(r.URL.Path, path) {
			if wl, ok := method(methodMap, r.Method); ok {
				return wl, ok
			}
		}
	}

	// 3. Try '*' path match.
	if methodMap, ok := cf.endpoints[allowAll]; ok {
		if wl, ok := method(methodMap, r.Method); ok {
			return wl, ok
		}
	}

	// 4. Not found.
	return whitelist{}, false
}

// method returns the whitelist for the specified HTTP method.
func method(methodMap map[string]whitelist, method string) (whitelist, bool) {
	// All methods in the map are lower-case.
	method = strings.ToLower(method)

	// 1. Try exact method match.
	if wl, ok := methodMap[method]; ok {
		return wl, ok
	}

	// 2. Try '*' method match.
	if wl, ok := methodMap[allowAll]; ok {
		return wl, ok
	}

	// 3. Not found.
	return whitelist{}, false
}

type whitelist struct {
	Whitelist
	cache map[string]struct{}
}

func wrapWhitelist(wl Whitelist) whitelist {
	cache := make(map[string]struct{}, len(wl))
	for _, v := range wl {
		cache[v] = struct{}{}
	}
	return whitelist{Whitelist: wl, cache: cache}
}

func (wl whitelist) contains(subject string) bool {
	_, ok := wl.cache[subject]
	return ok
}

func (wl whitelist) empty() bool {
	return len(wl.cache) == 0
}
