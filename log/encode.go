package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"unicode/utf16"
	"unicode/utf8"
)

// encode encodes and truncates value. See the package documentation for
// details about the process.
func encode(value string, max int) (encoded interface{}, truncated bool) {
	// Keep buf initially empty and unallocated. Defer writing of unescaped
	// characters until an escaped character is encountered. If no such
	// character is found, then the value will be returned unmodified (sans
	// truncation). Even if an escaped character is found, copying
	// unescaped characters in slices is better than writing byte-by-byte.
	var buf bytes.Buffer

	free := max // Free (unreserved) space left in buf.
	mark := 0   // Index of first deferred byte in value.
	for i, r := range value {
		// Unescaped single-byte characters.
		if free < 1 {
			value, truncated = value[:i], true
			break
		}
		if 0x20 <= r && r <= 0x7e && r != '"' && r != '\\' {
			// Reserve space in buf, but write later in slices (or
			// not at all if there are no escaped characters).
			free--
			continue
		}

		// r is an escaped character: ensure buf is initialized with a
		// quotation mark and copy deferred bytes before encoding r.
		if buf.Len() == 0 {
			buf.WriteByte('"')
		}
		buf.WriteString(value[mark:i])
		mark = i

		// Two character short encoding of common characters.
		if free < 2 {
			value, truncated = value[:i], true
			break
		}
		if short := escapeShort(r); short != "" {
			buf.WriteString(short)
			free -= 2
			mark++ // Only single byte characters have short encodings.
			continue
		}

		// UTF-16 encoding of Basic Multilingual Plane.
		if free < 6 {
			value, truncated = value[:i], true
			break
		}
		if r < 0x10000 {
			fmt.Fprintf(&buf, `\u%04x`, r)
			free -= 6
			size := utf8.RuneLen(r)
			if r == utf8.RuneError {
				// Check if a literal U+FFFD (size 3) or the
				// input contained bad UTF-8 (size 1).
				_, size = utf8.DecodeRuneInString(value[i:])
			}
			mark += size
			continue
		}

		// UTF-16 surrogate pair for supplementary planes.
		if free < 12 {
			value, truncated = value[:i], true
			break
		}
		r1, r2 := utf16.EncodeRune(r)
		fmt.Fprintf(&buf, `\u%04x\u%04x`, r1, r2)
		free -= 12
		mark += utf8.RuneLen(r)
	}

	// No encoding was needed: return (possibly) truncated value.
	if mark == 0 {
		return value, truncated
	}

	// Copy remaining deferred bytes and return buf as raw json message.
	buf.WriteString(value[mark:])
	buf.WriteByte('"')
	return json.RawMessage(buf.Bytes()), truncated
}

func escapeShort(r rune) string {
	switch r {
	case '"':
		return `\"`
	case '\\':
		return `\\`
	case '\b':
		return `\b`
	case '\f':
		return `\f`
	case '\n':
		return `\n`
	case '\r':
		return `\r`
	case '\t':
		return `\t`
	default:
		return ""
	}
}
