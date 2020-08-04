package confutil

import (
	"crypto/x509"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// nolint: lll, test data lines can be excessively long.
const (
	// JSON-escaped PEM-encoding of a self-signed X.509 certificate.
	testSelfSigned = `-----BEGIN CERTIFICATE-----\nMIIBfzCCASWgAwIBAgIUAUV5TeO0pFHNaBdcB2UFK1S12BgwCgYIKoZIzj0EAwIwFDESMBAGA1UEAwwJbG9jYWxob3N0MCAXDTE5MDYxMTExMTE0MVoYDzIxMTkwNTE4MTExMTQxWjAUMRIwEAYDVQQDDAlsb2NhbGhvc3QwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAASZvEFGMVXC5LR9TUlmbuT4OO1EWfkQsaIb55C/6mm/lZiexHe/J8vCAVkcPPCcj6w2kWXdz6czK2cymEHJPzYbo1MwUTAdBgNVHQ4EFgQUbnzZSGK3y7eYijP7v11jyGrRqtYwHwYDVR0jBBgwFoAUbnzZSGK3y7eYijP7v11jyGrRqtYwDwYDVR0TAQH/BAUwAwEB/zAKBggqhkjOPQQDAgNIADBFAiADfYyk0J38ZjAN2yHxpkB6nP8RrtfR9cXeZcLvlRFRPQIhAItQi8erHpIu0+GzGmKqUhHigbb6jQEeRuD8+0YNRUNV\n-----END CERTIFICATE-----`

	// JSON-escaped PEM-encoding of a X.509 CA certificate.
	testCA = `-----BEGIN CERTIFICATE-----\nMIIBbzCCARWgAwIBAgIUPG9Jx7zSC2v7fBrFCOodPzDUtBgwCgYIKoZIzj0EAwIw\nDTELMAkGA1UEAwwCY2EwHhcNMTkwNjI3MDgyMjU4WhcNMTkwNzI3MDgyMjU4WjAN\nMQswCQYDVQQDDAJjYTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABJm8QUYxVcLk\ntH1NSWZu5Pg47URZ+RCxohvnkL/qab+VmJ7Ed78ny8IBWRw88JyPrDaRZd3PpzMr\nZzKYQck/NhujUzBRMB0GA1UdDgQWBBRufNlIYrfLt5iKM/u/XWPIatGq1jAfBgNV\nHSMEGDAWgBRufNlIYrfLt5iKM/u/XWPIatGq1jAPBgNVHRMBAf8EBTADAQH/MAoG\nCCqGSM49BAMCA0gAMEUCIQDC0F+RKEX2deqAEKMQyc6j6CGJSBgJ+XQfore08ryq\nyAIgRSbsS8EluIfX/rhzIuAHTLbYEb0StkVIUY1soFG6wR0=\n-----END CERTIFICATE-----`

	// JSON-escaped PEM-encoding of a X.509 leaf certificate, signed by
	// testCA.
	testLeaf = `-----BEGIN CERTIFICATE-----\nMIIBGzCBwgIUcvlCmOI0x96c6Bwy+Dcl8A8va5QwCgYIKoZIzj0EAwIwDTELMAkG\nA1UEAwwCY2EwHhcNMTkwNjI3MDgyNTU1WhcNMTkwNzI3MDgyNTU1WjAUMRIwEAYD\nVQQDDAlsb2NhbGhvc3QwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAASZvEFGMVXC\n5LR9TUlmbuT4OO1EWfkQsaIb55C/6mm/lZiexHe/J8vCAVkcPPCcj6w2kWXdz6cz\nK2cymEHJPzYbMAoGCCqGSM49BAMCA0gAMEUCIQDnb9mBYyWZXzw1lur8Q5Gn1Ut3\nJE0MbcYpe5gc2O1b+gIgNrr53+N8adHB1wOcxp6bEtcqeN/sAUBhfDjRURDpFDA=\n-----END CERTIFICATE-----`

	// JSON-escaped PEM-encoding of the PKCS #8 private key of
	// testSelfSigned, testCA, and testLeaf.
	testPrivateKey = `-----BEGIN PRIVATE KEY-----\nMIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgDCq340CSzc+kFkzuRfQbrONNtm+nJps+mCSV1cOJaK6hRANCAASZvEFGMVXC5LR9TUlmbuT4OO1EWfkQsaIb55C/6mm/lZiexHe/J8vCAVkcPPCcj6w2kWXdz6czK2cymEHJPzYb\n-----END PRIVATE KEY-----`
)

func TestTLSUnmarshalJSON_LeafAndKey_LeafCertificate(t *testing.T) {
	const data = `{
		"chain": "` + testSelfSigned + `",
		"key": "` + testPrivateKey + `"
	}`
	var tls TLS
	if err := json.Unmarshal([]byte(data), &tls); err != nil {
		t.Fatal("unmarshal error:", err)
	}

	leaf, err := x509.ParseCertificate(tls.Certificate[0])
	if err != nil {
		t.Fatal("bad certificate error:", err)
	}
	if leaf.Subject.CommonName != "localhost" {
		t.Errorf("unexpected subject: got %s, expected localhost",
			leaf.Subject.CommonName)
	}
}

func TestTLSUnmarshalJSON_ChainAndKey_LeafCertificateWithCA(t *testing.T) {
	const data = `{
		"chain": "` + testLeaf + `\n` + testCA + `",
		"key": "` + testPrivateKey + `"
	}`
	var tls TLS
	if err := json.Unmarshal([]byte(data), &tls); err != nil {
		t.Fatal("unmarshal error:", err)
	}

	leaf, err := x509.ParseCertificate(tls.Certificate[0])
	if err != nil {
		t.Fatal("bad certificate error:", err)
	}
	if leaf.Subject.CommonName != "localhost" {
		t.Errorf("unexpected subject: got %s, expected localhost",
			leaf.Subject.CommonName)
	}

	ca, err := x509.ParseCertificate(tls.Certificate[1])
	if err != nil {
		t.Fatal("bad CA certificate error:", err)
	}
	if ca.Subject.CommonName != "ca" {
		t.Errorf("unexpected subject: got %s, expected ca",
			ca.Subject.CommonName)
	}
}

func TestCertificateUnmarshalJSON_PEMCertificate_X509Certificate(t *testing.T) {
	var cert Certificate
	if err := json.Unmarshal([]byte(`"`+testSelfSigned+`"`), &cert); err != nil {
		t.Fatal(err)
	}
	if cert.Subject.CommonName != "localhost" {
		t.Errorf("unexpected subject: got %s, expected localhost",
			cert.Subject.CommonName)
	}
}

func TestCertificateUnmarshalJSON_BadInput_Error(t *testing.T) {
	wedge := strings.Index(testSelfSigned, `\n`)
	withHeader := `"` + testSelfSigned[:wedge] + `\nHeader: value\n` + testSelfSigned[wedge:] + `"`

	tests := []struct {
		name  string
		input string
	}{
		{"non-PEM string", `"localhost"`},
		{"PRIVATE KEY", `"` + testPrivateKey + `"`},
		{"PEM header", withHeader},
		{"trailing data", `"` + testLeaf + `\n` + testCA + `"`},
	}

	var cert Certificate
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := json.Unmarshal([]byte(test.input), &cert)
			t.Log(err)
			if err == nil {
				t.Error("unexpected success")
			}
		})
	}
}

func TestCertPoolUnmarshalJSON_PEMArray_X509CertPool(t *testing.T) {
	const data = `[
		"` + testSelfSigned + `",
		"` + testCA + `"
	]`
	var pool CertPool
	if err := json.Unmarshal([]byte(data), &pool); err != nil {
		t.Fatal(err)
	}
	if size := len((*x509.CertPool)(&pool).Subjects()); size != 2 {
		t.Errorf("unexpected pool size: got %d, expected 2", size)
	}
}

func TestURLUnmarshalJSON_AbsoluteURL_ParsedURL(t *testing.T) {
	const data = "https://example.com/path"
	var url URL
	if err := json.Unmarshal([]byte(`"`+data+`"`), &url); err != nil {
		t.Fatal(err)
	}
	if url.Raw != data {
		t.Errorf("unexpected raw URL: got %s, expected %s", url.Raw, data)
	}
	if parsed := url.URL.String(); parsed != data {
		t.Errorf("unexpected parsed URL: got %s, expected %s", parsed, data)
	}
}

func TestURLUnmarshalJSON_EmptyString_Nil(t *testing.T) {
	var url URL
	if err := json.Unmarshal([]byte(`""`), &url); err != nil {
		t.Fatal(err)
	}
	if url.URL != nil {
		t.Error("unexpected URL from empty string:", url.URL)
	}
}

func TestURLUnmarshalJSON_RelativeURL_Error(t *testing.T) {
	var url URL
	if err := json.Unmarshal([]byte(`"/path"`), &url); err == nil {
		t.Error("unmarshaling relative URL succeeded")
	}
}

func TestURLUnmarshalJSON_OpaqueURL_Error(t *testing.T) {
	var url URL
	if err := json.Unmarshal([]byte(`"mailto:root@example.com"`), &url); err == nil {
		t.Error("unmarshaling opaque URL succeeded")
	}
}

func TestSecondsOr_Initialized_ReturnsValue(t *testing.T) {
	seconds := Seconds(2 * time.Second)
	if or := seconds.Or(3 * time.Second); or != time.Duration(seconds) {
		t.Error("non-0 seconds or default did not return seconds")
	}
}

func TestSecondsOr_Uninitialized_ReturnsDefault(t *testing.T) {
	var seconds Seconds
	def := 3 * time.Second
	if or := seconds.Or(def); or != def {
		t.Error("0 or default did not return default")
	}
}

func TestSecondsUnmarshalJSON_Number_DurationInSeconds(t *testing.T) {
	var seconds Seconds
	if err := json.Unmarshal([]byte{'5'}, &seconds); err != nil {
		t.Fatal(err)
	}
	if s := time.Duration(seconds); s != 5*time.Second {
		t.Errorf("unexpected value: got %s, expected 5s", s)
	}
}

func TestDateUnmarshalJSON_ValidString_ValidDate(t *testing.T) {
	var date Date
	if err := json.Unmarshal([]byte(`"2017-05-30"`), &date); err != nil {
		t.Fatal(err)
	}
	if date.Year() != 2017 || date.Month() != 5 || date.Day() != 30 {
		t.Errorf("unexpected value: got %s, expected 2017-05-30", date)
	} else {
		dts, dto := date.Zone()
		ns, no := time.Date(2019, 07, 01, 15, 40, 00, 00, time.Local).Zone()
		if dts != ns || dto != no {
			t.Errorf("unexpected time zone: got %s (%d), expected %s (%d)", dts, dto, ns, no)
		}
	}
}

func TestDateUnmarshalJSON_EmptyString_ZeroTime(t *testing.T) {
	var date Date
	if err := json.Unmarshal([]byte(`""`), &date); err != nil {
		t.Fatal(err)
	}
	if !date.IsZero() {
		t.Errorf("unexpected value: got %s, expected zero time", date)
	}
}

func TestDateUnmarshalJSON_InvalidString_Error(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"random text", `"anything"`},
		{"invalid date format", `"24 Jul 2000"`},
		{"date with time", `"` + time.Now().String() + `"`},
	}

	var date Date
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := json.Unmarshal([]byte(test.input), &date)
			t.Log(err)
			if err == nil {
				t.Error("unexpected success")
			}
		})
	}
}

func TestDateMarshalJSON_Date_ValidString(t *testing.T) {
	dateString := "2017-05-30"
	dt, err := time.ParseInLocation(dateLayout, dateString, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	date := Date{dt}
	bytes, err := json.Marshal(&date)
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes) != `"2017-05-30"` {
		t.Errorf("unexpected value: got %s, expected \"%s\"", string(bytes), dateString)
	}
}

func TestDateMarshalJSON_UTCDate_ValidString(t *testing.T) {
	dateString := "2017-05-30"
	dt, err := time.ParseInLocation(dateLayout, dateString, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	date := Date{dt.In(time.UTC)}
	bytes, err := json.Marshal(&date)
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes) != `"2017-05-30"` {
		t.Errorf("unexpected value: got %s, expected \"%s\"", string(bytes), dateString)
	}
}

func TestDateMarshalJSON_ZeroTime_EmptyString(t *testing.T) {
	date := Date{time.Time{}}
	bytes, err := json.Marshal(&date)
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes) != `""` {
		t.Errorf("unexpected value: got %s, expected \"\"", string(bytes))
	}
}

func TestTimeUnmarshalJSON_ValidString_ValidTime(t *testing.T) {
	var tt Time
	if err := json.Unmarshal([]byte(`"09:52"`), &tt); err != nil {
		t.Fatal(err)
	}
	if tt.Hour() != 9 || tt.Minute() != 52 {
		t.Errorf("unexpected value: got %s, expected 09:52", tt)
	}
}

func TestTimeUnmarshalJSON_MidnightString_NonZeroTime(t *testing.T) {
	var tt Time
	if err := json.Unmarshal([]byte(`"00:00"`), &tt); err != nil {
		t.Fatal(err)
	}
	if tt.IsZero() {
		t.Error("unexpected zero time")
	}
}

func TestTimeUnmarshalJSON_EmptyString_ZeroTime(t *testing.T) {
	var tt Time
	if err := json.Unmarshal([]byte(`""`), &tt); err != nil {
		t.Fatal(err)
	}
	if !tt.IsZero() {
		t.Errorf("unexpected value: got %s, expected zero time", tt)
	}
}

func TestTimeUnmarshalJSON_InvalidString_Error(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"random text", `"anything"`},
		{"invalid time format", `"09.52"`},
		{"time with seconds", `"09:52:00"`},
		{"date with time", `"` + time.Now().String() + `"`},
	}

	var tt Time
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := json.Unmarshal([]byte(test.input), &tt)
			t.Log(err)
			if err == nil {
				t.Error("unexpected success")
			}
		})
	}
}

func TestTimeMarshalJSON_NonZeroTime_ValidString(t *testing.T) {
	str := "09:52"
	tt, err := ParseTime(str)
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := json.Marshal(tt)
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes) != `"` + str + `"` {
		t.Errorf("unexpected value: got %s, expected \"%s\"", string(bytes), str)
	}
}

func TestTimeMarshalJSON_MidnightTime_ValidString(t *testing.T) {
	str := "00:00"
	tt, err := ParseTime(str)
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := json.Marshal(tt)
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes) != `"` + str + `"` {
		t.Errorf("unexpected value: got %s, expected \"%s\"", string(bytes), str)
	}
}

func TestTimeMarshalJSON_ZeroTime_EmptyString(t *testing.T) {
	var tt Time
	bytes, err := json.Marshal(tt)
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes) != `""` {
		t.Errorf("unexpected value: got %s, expected \"\"", string(bytes))
	}
}

func TestDateTimeMarshalJSON_TimeWithTZ_ValidString(t *testing.T) {
	str := "2020-03-06T11:37:45+02:00"
	dt, err := time.Parse(time.RFC3339, str)
	if err != nil {
		t.Fatal(err)
	}
	datetime := DateTime{Time: dt}
	bytes, err := json.Marshal(datetime)
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes) != `"`+str+`"` {
		t.Errorf("unexpected value: got %s, expected \"%s\"", string(bytes), str)
	}
}

func TestDateTimeMarshalJSON_ZeroTime_EmptyString(t *testing.T) {
	bytes, err := json.Marshal(DateTime{})
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes) != `""` {
		t.Errorf("unexpected value: got %s, expected \"\"", string(bytes))
	}
}

func TestDateTimeUnmarshalJSON_EmptyString_ZeroTime(t *testing.T) {
	var date DateTime
	if err := json.Unmarshal([]byte(`""`), &date); err != nil {
		t.Fatal(err)
	}
	if !date.IsZero() {
		t.Errorf("unexpected value: got %s, expected zero time", date)
	}
}

func TestDurationUnmarshalJSON_ValidString_ValidDuration(t *testing.T) {
	// given
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{`"3600h"`, 3600 * time.Hour},
		{`"24h"`, 24 * time.Hour},
		{`"1h30m"`, 1*time.Hour + 30*time.Minute},
		{`"30m1h"`, 30*time.Minute + 1*time.Hour},
		{`"90m"`, 90 * time.Minute},
		{`"60m30m"`, 60*time.Minute + 30*time.Minute},
		{`"90000ms"`, 90000 * time.Millisecond},
		{`""`, time.Duration(0)},
		{`null`, time.Duration(0)},
	}

	// loop - when/then
	var d Duration
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			// when
			err := json.Unmarshal([]byte(test.input), &d)
			// then
			if err != nil {
				t.Error("unexpected error", err)
			}
			if d.Duration != test.expected {
				t.Errorf("unexpected result: got %v, expected %v", d, test.expected)
			}
		})
	}
}

func TestDurationUnmarshalJSON_InvalidString_Error(t *testing.T) {
	// given
	tests := []struct {
		name  string
		input string
	}{
		{"no unit", `"3600"`},
		{"unknown unit", `"24t"`},
		{"other text", `"hey 1h30m"`},
		{"spaces", `"90 000ms"`},
	}

	// loop - when/then
	var d Duration
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// when
			err := json.Unmarshal([]byte(test.input), &d)
			// then
			t.Log(err)
			if err == nil {
				t.Error("unexpected success")
			}
		})
	}
}

func TestDurationMarshalJSON_Duration_ValidString___(t *testing.T) {
	// given
	d := Duration{2*time.Hour + 30*time.Minute}
	// when
	bytes, err := json.Marshal(&d)
	// then
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes) != `"2h30m0s"` {
		t.Errorf("unexpected value: got %s, expected \"%s\"", string(bytes), "2h30m0s")
	}
}

func TestDurationMarshalJSON_Duration_ValidString(t *testing.T) {
	// given
	tests := []struct {
		expected string
		input    time.Duration
	}{
		{`"3600h0m0s"`, 3600 * time.Hour},
		{`"24h0m0s"`, 24 * time.Hour},
		{`"1h30m0s"`, 1*time.Hour + 30*time.Minute},
		{`"1m30s"`, 1*time.Minute + 30*time.Second},
		{`""`, time.Duration(0)},
	}

	// loop - when/then
	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			// when
			bytes, err := json.Marshal(&Duration{test.input})
			// then
			if err != nil {
				t.Fatal(err)
			}
			if string(bytes) != test.expected {
				t.Errorf("unexpected value: got %s, expected \"%s\"", string(bytes), test.expected)
			}
		})
	}
}
