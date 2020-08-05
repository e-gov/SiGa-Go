package httpsutil

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"stash.ria.ee/vis3/vis3-common/pkg/confutil"
	"stash.ria.ee/vis3/vis3-common/pkg/errorutil"
	"stash.ria.ee/vis3/vis3-common/pkg/log"
	"stash.ria.ee/vis3/vis3-common/pkg/validation"
)

// Parser is a helper for parsing HTTP request parameters.
//
// If any of the Parser parsing methods encounter errors, then these are
// appended to an internal error list and undefined values are returned. It is
// therefore necessary to consult Err before using the values returned by
// parsing methods.
type Parser struct {
	// Path is a map of path parameters. When parsing path parameters, then
	// the keys are considered mandatory and missing values are errors.
	Path map[string]string

	// Query is the query string of the request. When parsing query
	// parameters, then they are considered optional (i.e., when they are
	// missing, a zero value is returned and no error is appended), but
	// must not have more than a single value.
	Query url.Values

	// errors is the slice of errors encountered during parsing.
	errors errorutil.Slice

	// queried tracks the query parameters that have been parsed.
	queried map[string]struct{}
}

// PathInt parses the named path parameter as an integer.
func (p *Parser) PathInt(key string) int {
	v, err := strconv.Atoi(p.Path[key])
	if err != nil {
		p.errors = p.errors.Append(validation.InvalidValue(key, p.Path[key]))
		return 0
	}
	return v
}

func (p *Parser) addQueried(key string) {
	if p.queried == nil {
		p.queried = make(map[string]struct{})
	}
	p.queried[key] = struct{}{}
}

func (p *Parser) query(key string) (string, bool) {
	p.addQueried(key)
	switch vs := p.Query[key]; len(vs) {
	case 0:
	case 1:
		return vs[0], true
	default:
		p.errors = p.errors.Append(validation.InvalidValue(key, vs))
	}
	return "", false
}

// QueryString returns the named query parameter. QueryString is preferred over
// directly accessing the key so that Parser can check for duplicates and mark
// it as queried.
func (p *Parser) QueryString(key string) string {
	str, _ := p.query(key)
	return str
}

// QueryStrings parses the named query parameter as a list of strings separated
// with commas (','). Currently there is no way to escape a literal comma.
func (p *Parser) QueryStrings(key string) []string {
	if str, ok := p.query(key); ok {
		return strings.Split(str, ",")
	}
	return nil
}

// QueryInt parses the named query parameter as an integer.
func (p *Parser) QueryInt(key string) int {
	if str, ok := p.query(key); ok {
		v, err := strconv.Atoi(str)
		if err == nil {
			return v
		}
		p.errors = p.errors.Append(validation.InvalidValue(key, str))
	}
	return 0
}

func (p *Parser) QueryUint(key string) uint {
	if str, ok := p.query(key); ok {
		v, err := strconv.Atoi(str)
		if err == nil {
			return uint(v)
		}
		p.errors = p.errors.Append(validation.InvalidValue(key, str))
	}
	return 0
}

// QueryInts parses the named query parameter as a list of integers separated
// with commas (','). Currently there is no way to escape a literal comma.
func (p *Parser) QueryInts(key string) []int {
	var ints []int
	for _, str := range p.QueryStrings(key) {
		v, err := strconv.Atoi(str)
		if err != nil {
			p.errors = p.errors.Append(validation.InvalidValue(key, str))
			continue
		}
		ints = append(ints, v)
	}
	return ints
}

// QueryDate parses the named query parameter as a date.
func (p *Parser) QueryDate(key string) confutil.Date {
	if str, ok := p.query(key); ok {
		v, err := confutil.ParseDate(str)
		if err == nil {
			return v
		}
		p.errors = p.errors.Append(validation.InvalidValue(key, str))
	}
	return confutil.Date{}
}

// QueryDateTime parses the named query parameter as a dateTime.
func (p *Parser) QueryDateTime(key string) confutil.DateTime {
	if str, ok := p.query(key); ok {
		v, err := confutil.ParseDateTime(str)
		if err == nil {
			return v
		}
		p.errors = p.errors.Append(validation.InvalidValue(key, str))
	}
	return confutil.DateTime{}
}

// QueryBool parses the named query parameter as a boolean.
func (p *Parser) QueryBool(key string) bool {
	if str, ok := p.query(key); ok {
		v, err := strconv.ParseBool(str)
		if err == nil {
			return v
		}
		p.errors = p.errors.Append(validation.InvalidValue(key, str))
	}
	return false
}

// QueryMap returns the values of the named query parameters in a map.
func (p *Parser) QueryMap(keys ...string) map[string]string {
	m := make(map[string]string, len(keys))
	for _, key := range keys {
		if str, ok := p.query(key); ok {
			m[key] = str
		}
	}
	return m
}

// Err returns the list of errors encountered during parsing.
//
// It also checks if p.Query has some extra parameters that were not parsed. If
// so, then these are added to the error list.
func (p *Parser) Err() error {
	extended := p.errors
	for key := range p.Query {
		if _, ok := p.queried[key]; !ok {
			extended = extended.Append(validation.InvalidValue(key, p.Query.Get(key)))
		}
	}
	return extended.AsError()
}

// UnmarshalJSON unmarshals a JSON-encoded HTTP request body into v. In case of
// errors, it logs any details and responds to w with http.StatusBadRequest.
// The returned boolean indicates success.
func UnmarshalJSON(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		log.Error().WithError(errors.WithStack(err)).Skip(1).
			Log(r.Context(), "unmarshal_json_error")
		w.WriteHeader(http.StatusBadRequest)
		return false
	}
	return true
}
