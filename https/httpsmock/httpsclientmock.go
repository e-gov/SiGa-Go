/*
Package httpsmock provides a mock implementation of the https.Client interface,
which returns preconfigured values.
*/
package httpsmock

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/pkg/errors"

	"stash.ria.ee/vis3/vis3-common/pkg/https"
)

// Client implements the https.Client interface.
type Client struct {
	// ExpectedPath is the expected URL path hit by the test
	ExpectedPath string
	// Bytes is the expected mock response that is given to the Req callback.
	Bytes []byte
	// Error is the expected error returned by the Do method.
	Error error
	// realPath is the real path hit by the test
	realPath string
}

// WithJSONResponse sets the response bytes JSON-encoding the given object.
func (c *Client) WithJSONResponse(resp interface{}) *Client {
	var err error
	if c.Bytes, err = json.Marshal(resp); err != nil {
		panic(err)
	}
	return c
}

// Req implements the https.Client interface.
func (c *Client) Req(ctx context.Context, parts ...string) *https.Req {
	c.realPath = path.Join(parts...)
	return &https.Req{
		Client: c,
	}
}

// Do implements the https.Client interface.
func (c *Client) Do(r *https.Req) error {
	if c.ExpectedPath != "" && c.realPath != c.ExpectedPath {
		return errors.Errorf("Unexpected path: got %s expected %s", c.realPath, c.ExpectedPath)
	}
	if c.Error != nil {
		return c.Error
	}
	return r.Handle(&http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(c.Bytes)),
	})
}
