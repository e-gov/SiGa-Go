package siga

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/e-gov/SiGa-Go/https"
	"github.com/e-gov/SiGa-Go/https/httpsutil"
	"github.com/pkg/errors"
)

// httpClient on SiGa poole pöörduv HTTPS klient.
type httpClient struct {
	client *http.Client

	url        string
	identifier string
	key        []byte
	algo       string
	hmac       func() hash.Hash
	now        func() time.Time
}

// newHTTPClient moodustab conf põhjal SiGa kliendi.
func newHTTPClient(conf Conf) (*httpClient, error) {
	c := &httpClient{
		client: &http.Client{
			Transport: httpsutil.Transport(conf.RootCAs, &conf.ClientTLS),
			Timeout:   conf.Timeout.Or(https.DefaultClientTimeout),
		},
		url:        conf.URL.Raw,
		identifier: conf.ServiceIdentifier,
		key:        []byte(conf.ServiceKey),
		now:        time.Now,
	}

	switch conf.HMACAlgorithm {
	case "", "HMAC-SHA256":
		c.algo = "HmacSHA256"
		c.hmac = sha256.New
	case "HMAC-SHA384":
		c.algo = "HmacSHA384"
		c.hmac = sha512.New384
	case "HMAC-SHA512":
		c.algo = "HmacSHA512"
		c.hmac = sha512.New
	default:
		return nil, errors.Errorf("unknown HMACAlgorithm: %s", conf.HMACAlgorithm)
	}
	return c, nil
}

// authHeaders seab SiGa kliendile autoriseerimispäised.
func (c *httpClient) authHeaders(headers http.Header, method, uri string, body []byte) {
	now := strconv.FormatInt(c.now().Unix(), 10)
	hmac := hmac.New(c.hmac, c.key)

	fmt.Fprintf(hmac, "%s:%s:%s:%s:", c.identifier, now, method, uri)
	hmac.Write(body)

	headers.Set("X-Authorization-Timestamp", now)
	headers.Set("X-Authorization-ServiceUUID", c.identifier)
	headers.Set("X-Authorization-Hmac-Algorithm", c.algo)
	headers.Set("X-Authorization-Signature", hex.EncodeToString(hmac.Sum(nil)))
}

// do täidab SiGa kliendina päringu.
func (c *httpClient) do(ctx context.Context, method, uri string, req interface{}, resp interface{}) error {
	// If a request body is given, then marshal it into memory since we
	// need to calculate the MAC over it before sending it to the server.
	var body []byte
	var bodyReader io.Reader
	if req != nil {
		var err error
		if body, err = json.Marshal(req); err != nil {
			return errors.Wrap(err, "encode request")
		}
		bodyReader = bytes.NewReader(body)
	}

	// Create a HTTP request and set the required headers.
	httpReq, err := http.NewRequest(method, singleJoiningSlash(c.url, uri), bodyReader)
	if err != nil {
		return errors.WithStack(err)
	}
	httpReq = httpReq.WithContext(ctx)
	if req != nil {
		httpReq.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}
	c.authHeaders(httpReq.Header, method, uri, body)

	// Perform the request.
	// log.Debug().
	// 	WithString("method", method).
	// 	WithString("url", httpReq.URL).
	// 	WithString("contentLength", httpReq.ContentLength).
	// 	Log(ctx, "request")
	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return errors.Wrap(err, "perform request")
	}
	defer httpResp.Body.Close()
	// log.Debug().WithString("status", httpResp.StatusCode).Log(ctx, "response")

	// Decode the response body into resp or an error struct if the HTTP
	// status code indicates failure.
	//
	// TODO: Apply an upper limit to the number of bytes read. For now
	// trust the service provider to not return excessively large bodies.
	decoder := json.NewDecoder(httpResp.Body)
	if httpResp.StatusCode/100 != 2 { // XXX: Exact codes?
		errResp := errorResponse{statusCode: httpResp.StatusCode}
		if httpResp.Body != http.NoBody {
			if err := decoder.Decode(&errResp); err != nil {
				errResp.decodeErr = err
			}
		}
		return errors.WithStack(errResp)
	}
	if resp != nil {
		return errors.Wrap(decoder.Decode(resp), "decode response")
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

type errorResponse struct {
	statusCode   int
	decodeErr    error
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

func (e errorResponse) Error() string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "service error: http %d", e.statusCode)
	if e.decodeErr != nil {
		fmt.Fprintf(&buf, ", decode err: %v", e.decodeErr)
	} else if e.ErrorCode != "" {
		fmt.Fprintf(&buf, ", code %s, %s", e.ErrorCode, e.ErrorMessage)
	}
	return buf.String()
}
