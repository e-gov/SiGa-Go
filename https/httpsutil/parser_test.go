package httpsutil

import (
	"net/url"
	"testing"
)

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
