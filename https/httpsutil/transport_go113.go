// +build go1.13

package httpsutil

import "net/http"

func cloneDefaultTransport() *http.Transport {
	return http.DefaultTransport.(*http.Transport).Clone()
}
