package log

import (
	"encoding/json"
	"testing"
)

func TestEncode_Value_EncodedAndTruncated(t *testing.T) {
	tests := []struct {
		name      string // If empty, then value is used.
		value     string
		max       int // MaxParameterLength if zero.
		output    string
		truncated bool
	}{
		{
			name:   "empty",
			value:  "",
			output: `""`,
		},
		{
			value:  "plain",
			output: `"plain"`,
		},
		{
			name:   "control",
			value:  "\"\\\b\f\n\r\t",
			output: `"\"\\\b\f\n\r\t"`,
		},
		{
			value:  "šžõäöü",
			output: `"\u0161\u017e\u00f5\u00e4\u00f6\u00fc"`,
		},
		{
			name:   "supplementary",
			value:  "👍", // U+1F44D
			output: `"\ud83d\udc4d"`,
		},
		{
			name:   "replacement",
			value:  "head\ufffdtail",
			output: `"head\ufffdtail"`,
		},
		{
			name:   "invalid utf8",
			value:  "head\xfftail",
			output: `"head\ufffdtail"`,
		},
		{
			value:     "truncated",
			max:       5,
			output:    `"trunc"`,
			truncated: true,
		},
		{
			name:      "truncated post-escape",
			value:     "head\ntail",
			max:       8,
			output:    `"head\nta"`,
			truncated: true,
		},
		{
			name:      "truncated control",
			value:     "linefeed=\n",
			max:       10,
			output:    `"linefeed="`,
			truncated: true,
		},
		{
			name:      "truncated unicode",
			value:     "o-with-tilde=õ",
			max:       18,
			output:    `"o-with-tilde="`,
			truncated: true,
		},
		{
			name:      "truncated surrogate pair",
			value:     "thumbs-up=👍",
			max:       21,
			output:    `"thumbs-up="`,
			truncated: true,
		},
	}

	for _, test := range tests {
		name := test.name
		if name == "" {
			name = test.value
		}
		t.Run(name, func(t *testing.T) {
			max := MaxParameterLength
			if test.max != 0 {
				max = test.max
			}
			encoded, truncated := encode(test.value, max)

			bytes, err := json.Marshal(encoded)
			if err != nil {
				t.Fatal(err)
			}
			if output := string(bytes); output != test.output {
				t.Errorf("unexpected encoding,\nencoded:  %s\nexpected: %s",
					output, test.output)
			}

			if truncated != test.truncated {
				t.Errorf("unexpected truncation,\ntruncated: %t\nexpected:  %t",
					truncated, test.truncated)
			}
		})
	}
}
