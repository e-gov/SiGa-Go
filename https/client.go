package https

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/pkg/errors"

	"github.com/e-gov/SiGa-Go/https/httpsutil"
)

// Callback is the callback function type.
type Callback func(*http.Response) error

// Client is a stateless HTTP/HTTPS client.
type Client interface {
	// Req combines a URL from the base url and the provided parts and returns a request builder.
	Req(ctx context.Context, parts ...string) *Req
	// Do executes the given pre-configured request.
	Do(req *Req) error
}

// client implements the https.Client interface.
type client struct {
	url    string
	client *http.Client
}

// NewClient creates a new HTTPS client for the specified configuration.
func NewClient(conf ClientConf) Client {
	return &client{
		url: conf.URL.Raw,
		client: &http.Client{
			Transport: httpsutil.Transport(conf.RootCAs, &conf.ClientTLS),
			Timeout:   conf.Timeout.Or(DefaultClientTimeout),
		},
	}
}

// Req implements the https.Client interface.
//
// Parts are joined together with '/'. It is ensured that single slash is used
// for appending parts to the configured base URL.
func (c *client) Req(ctx context.Context, parts ...string) *Req {
	url := singleJoiningSlash(c.url, path.Join(parts...))
	// log.Debug().WithString("url", url).Log(ctx, "new_request")
	return &Req{
		Client: c,
		ctx:    ctx,
		url:    url,
	}
}

// Do implements the https.Client interface.
func (c *client) Do(r *Req) error {
	// If an error was recorded setting up the request, just return it.
	if r.Err != nil {
		return r.Err
	}

	req, err := http.NewRequest(r.method, r.url, r.body)
	if err != nil {
		return err
	}
	if r.contentType != "" {
		req.Header.Set("Content-Type", r.contentType)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// log.Debug().
	// 	WithString("url", r.url).
	// 	WithStringf("status", "%d", res.StatusCode).
	// 	Log(r.ctx, "got_response")

	return r.Handle(res)
}

// Req is a helper to build and execute a HTTP requests.
type Req struct {
	Client      Client // Back reference to the https.Client.
	ctx         context.Context
	method      string    // The HTTP request method (GET, POST etc).
	url         string    // The URL of the request.
	contentType string    // The content type.
	body        io.Reader // The request body.
	successCB   Callback  // The callback for 2XX response status codes.
	clientErrCB Callback  // The callback for 4XX response status codes.
	responseCB  Callback  // The callback for any response.
	Err         error     // The error stored during build-up of the request.
	Status      int       // The HTTP response status code.
}

// WithJSON evaluates the request body and content type as a JSON data with the specified content.
func (r *Req) WithJSON(body interface{}) *Req {
	var b []byte
	if b, r.Err = json.Marshal(body); r.Err != nil {
		return r
	}
	r.body = bytes.NewReader(b)
	r.contentType = "application/json"
	return r
}

// WithXML evaluates the request body and content type as XML data with the specified content.
func (r *Req) WithXML(body interface{}) *Req {
	var b []byte
	if b, r.Err = xml.Marshal(body); r.Err != nil {
		return r
	}
	r.body = bytes.NewReader(b)
	r.contentType = "text/xml"
	return r
}

// OnSuccess registeres the callback for 2XX response codes.
func (r *Req) OnSuccess(cb Callback) *Req {
	r.successCB = cb
	return r
}

// OnClientError registeres the callback for 4XX response codes.
func (r *Req) OnClientError(cb Callback) *Req {
	r.clientErrCB = cb
	return r
}

// OnResponse registeres the callback for any response code.
func (r *Req) OnResponse(cb Callback) *Req {
	r.responseCB = cb
	return r
}

// GetJSON makes a GET request and parses the response as JSON into the given object.
func (r *Req) GetJSON(ret interface{}) error {
	return r.OnSuccess(func(res *http.Response) error {
		return json.NewDecoder(res.Body).Decode(ret)
	}).Get()
}

// Get makes a GET request using the given function as response handler.
func (r *Req) Get() error {
	r.method = http.MethodGet
	return r.Client.Do(r)
}

// Post makes a POST request using the given function as response handler.
func (r *Req) Post() error {
	r.method = http.MethodPost
	return r.Client.Do(r)
}

// Put makes a PUT request using the given function as response handler.
func (r *Req) Put() error {
	r.method = http.MethodPut
	return r.Client.Do(r)
}

// Delete makes a DELETE request using the given function as response handler.
func (r *Req) Delete() error {
	r.method = http.MethodDelete
	return r.Client.Do(r)
}

// Handle handles the HTTP response by calling the appropriate preconfigured callback.
func (r *Req) Handle(res *http.Response) error {
	r.Status = res.StatusCode
	respClass := res.StatusCode / 100

	// Successful response.
	if respClass == 2 && r.successCB != nil {
		return r.successCB(res)
	}

	// Client error response.
	if respClass == 4 && r.clientErrCB != nil {
		return r.clientErrCB(res)
	}

	// Any other response code.
	if r.responseCB != nil {
		return r.responseCB(res)
	}

	// The default handling: treat every response code other than 2xx as an error, discard the body.
	if respClass != 2 {
		return errors.Errorf("unexpected response status %d", res.StatusCode)
	}
	return nil
}

// singleJoiningSlash joins the given paths with single slash in between.
// Copied from net/http/httputil/reverseproxy.go.
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
