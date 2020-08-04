package siga

import (
	"crypto/sha256"
	"net/http"
	"testing"
	"time"
)

func TestHTTPClientAuthHeaders_WikiExample_Matches(t *testing.T) {
	// Test against the "reference" values from SiGa wiki at
	// https://github.com/open-eid/SiGa/wiki/Authorization.

	// given
	c := &httpClient{
		identifier: "a7fd7728-a3ea-4975-bfab-f240a67e894f",
		key:        []byte("746573745365637265744b6579303031"),
		algo:       "HmacSHA256",
		hmac:       sha256.New,
		now: func() time.Time {
			return time.Unix(1580400796, 0)
		},
	}
	headers := make(http.Header)
	method := http.MethodPost
	uri := "/hashcodecontainers"
	body := []byte(`{"dataFiles":[{"fileName":"test.txt","fileHashSha512":"hQVz9wirVZNvP/q3HoaW8nu0FfvrGkZinhADKE4Y4j/dUuGfgONfR4VYdu0p/dj/yGH0qlE0FGsmUB2N3oLuhA==","fileSize":189,"fileHashSha256":"RnKZobNWVy8u92sDL4S2j1BUzMT5qTgt6hm90TfAGRo="}]}`)

	// when
	c.authHeaders(headers, method, uri, body)

	// then
	assert := func(key, value string) {
		if got := headers.Get(key); got != value {
			t.Errorf("unexpected %s:\n     got: %s\nexpected: %s", key, got, value)
		}
	}
	assert("X-Authorization-Timestamp", "1580400796")
	assert("X-Authorization-ServiceUUID", "a7fd7728-a3ea-4975-bfab-f240a67e894f")
	assert("X-Authorization-Hmac-Algorithm", "HmacSHA256")
	assert("X-Authorization-Signature", "7301b3b88995b410bed0016b9a5bb3d177d32ac2bb2e91fabb80c084180eb42d")
}
