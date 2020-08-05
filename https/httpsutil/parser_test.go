package httpsutil

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"stash.ria.ee/vis3/vis3-common/pkg/confutil"
	"testing"
)

func TestParserQueryUInt_SingleValue(t *testing.T) {
	p := Parser{Query: url.Values{
		"id": []string{"1"},
	}}

	queryUint := p.QueryUint("id")
	assert.Nil(t, p.Err())
	assert.Equal(t, uint(1), queryUint)
}

func TestParserQueryString_MultipleValues_Error(t *testing.T) {
	// given
	p := Parser{Query: url.Values{
		"string": []string{"foo", "bar"},
	}}

	// when
	p.QueryString("string")
	err := p.Err()

	// then
	if err == nil || err.Error() != `invalid_value: code: invalid_value, field: string, value: [foo bar]` {
		t.Error("unexpected error:", err)
	}
}

func TestParserQueryStrings_ListValue_ReturnsMultipleValues(t *testing.T) {
	// given
	p := Parser{Query: url.Values{
		"list": []string{`foo,bar,`},
	}}

	// when
	list := p.QueryStrings("list")

	// then
	if err := p.Err(); err != nil {
		t.Error("unexpected error:", err)
	}
	if len(list) != 3 || list[0] != "foo" || list[1] != "bar" || list[2] != "" {
		t.Errorf("unexpected result: got %v, expected [foo bar ]", list)
	}
}

func TestParserQueryInts_ListValue_ReturnsMultipleValues(t *testing.T) {
	// given
	p := Parser{Query: url.Values{
		"id": []string{"1,2,3"},
	}}

	// when
	ints := p.QueryInts("id")

	// then
	if err := p.Err(); err != nil {
		t.Error("unexpected error:", err)
	}
	if len(ints) != 3 || ints[0] != 1 || ints[1] != 2 || ints[2] != 3 {
		t.Errorf("unexpected result: got %v, expected [1 2 3]", ints)
	}
}

func TestParserErr_UnparsedQueryParameter_Error(t *testing.T) {
	// given
	p := Parser{Query: url.Values{
		"queried": []string{"value"},
		"extra":   []string{"value"},
	}}

	// when
	p.QueryString("queried")
	err := p.Err()

	// then
	if err == nil || err.Error() != `invalid_value: code: invalid_value, field: extra, value: value` {
		t.Error("unexpected error:", err)
	}
}

func TestParserQueryDateTime_ok(t *testing.T) {
	dateTimeString := "2020-05-05T01:02:03Z"
	p := Parser{Query: url.Values{
		"dateTime": []string{dateTimeString},
	}}

	queryDateTime := p.QueryDateTime("dateTime")
	assert.Nil(t, p.Err())
	parsedDateTime, err := confutil.ParseDateTime(dateTimeString)
	assert.Nil(t, err)
	assert.Equal(t, parsedDateTime, queryDateTime)
}


func TestParserQueryDateTime_error(t *testing.T) {
	invalidDateTimeString := "x2020-05-05"
	p := Parser{Query: url.Values{
		"dateTime": []string{invalidDateTimeString},
	}}

	p.QueryDateTime("dateTime")
	assert.Error(t, p.Err())
}
