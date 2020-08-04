package https

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/e-gov/SiGa-Go/confutil"
)

func TestGetJSON_ServerRespondsOK_ResultReceived(t *testing.T) {
	// given
	expected := respdata{Value: "Server response"}
	s := newMockserver(http.StatusOK)
	defer s.server.Close()
	conf := ClientConf{
		URL: confutil.URL{
			Raw: s.server.URL,
		},
		RootCAs: s.rootCAs,
	}
	client := NewClient(conf)
	var result respdata
	// when
	err := client.Req(nil).WithJSON(expected).GetJSON(&result)
	// then
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if result.Method != http.MethodGet || result.Value != expected.Value {
		t.Errorf("unexpected result: got %+v, expected: %+v", result, expected)
	}
}

func TestPost_ServerRespondsNOK_ErrorReturned(t *testing.T) {
	// given
	s := newMockserver(http.StatusBadRequest)
	defer s.server.Close()
	conf := ClientConf{
		URL: confutil.URL{
			Raw: s.server.URL,
		},
		RootCAs: s.rootCAs,
	}
	client := NewClient(conf)
	// when
	err := client.Req(nil).Post()
	// then
	if err == nil {
		t.Errorf("unexpected success")
	}
}

// ---- Helper methods ----

// mockserver is mock HTTP server over TLS that counts requests and returns
// the request path to the caller.
type mockserver struct {
	server  *httptest.Server
	counter int
	rootCAs *confutil.CertPool
}

func newMockserver(status int) *mockserver {
	s := mockserver{}
	s.server = httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.counter++

			var resp respdata
			if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			resp.Method = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			json.NewEncoder(w).Encode(resp)
		}))
	s.rootCAs = (*confutil.CertPool)(s.server.Client().Transport.(*http.Transport).TLSClientConfig.RootCAs)
	return &s
}

type respdata struct {
	Method string
	Value  string
}
