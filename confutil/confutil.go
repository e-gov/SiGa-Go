/*
Package confutil contains utility types and functions which help with
configuration loading.
*/
package confutil

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04"
)

// TLS is a TLS certificate chain and private key.
type TLS tls.Certificate

// UnmarshalJSON unmarshals TLS from a {"chain": "string", "key": "string"}
// JSON object where
//
//   - "chain" contains the concatenated PEM-encodings of X.509 certificates
//     (leaf certificate first) and
//
//   - "key" the PEM-encoding of the PKCS #8 private key for the leaf.
//
func (t *TLS) UnmarshalJSON(data []byte) error {
	if noop(data) {
		return nil
	}

	var parsed struct {
		Chain string
		Key   string
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return errors.Wrap(err, "unmarshal TLS")
	}
	cert, err := tls.X509KeyPair([]byte(parsed.Chain), []byte(parsed.Key))
	if err != nil {
		return errors.Wrap(err, "parse TLS")
	}
	*t = TLS(cert)
	return nil
}

// Certificate is a single X.509 certificate.
type Certificate x509.Certificate

// UnmarshalJSON unmarshals Certificate from a JSON string containing the
// PEM-encoding of the X.509 certificate.
func (c *Certificate) UnmarshalJSON(data []byte) error {
	if noop(data) {
		return nil
	}

	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return errors.Wrap(err, "unmarshal certificate")
	}
	block, rest := pem.Decode([]byte(raw))
	if block == nil {
		return errors.New("certificate not PEM-encoded")
	}
	if block.Type != "CERTIFICATE" {
		return errors.Errorf("not a certificate: %s", block.Type)
	}
	if len(block.Headers) > 0 {
		return errors.New("PEM headers not allowed")
	}
	if len(rest) > 0 {
		return errors.Errorf("certificate has %d trailing bytes", len(rest))
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return errors.Wrap(err, "parse certificate")
	}
	*c = Certificate(*cert)
	return nil
}

// CertPool is a set of X.509 certificates.
type CertPool x509.CertPool

// UnmarshalJSON unmarshals CertPool from a JSON array of encoded Certificates.
func (p *CertPool) UnmarshalJSON(data []byte) error {
	if noop(data) {
		return nil
	}

	var certs []*Certificate
	if err := json.Unmarshal(data, &certs); err != nil {
		return errors.Wrap(err, "unmarshal certificates")
	}
	pool := x509.NewCertPool()
	for _, cert := range certs {
		pool.AddCert((*x509.Certificate)(cert))
	}
	*p = CertPool(*pool)
	return nil
}

// URL is an absolute URL. It has the original string and a parsed structure.
type URL struct {
	Raw string
	URL *url.URL
}

// UnmarshalJSON unmarshals URL from a JSON string. The URL must be absolute.
func (u *URL) UnmarshalJSON(data []byte) error {
	if noop(data) {
		return nil
	}

	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return errors.Wrap(err, "unmarshal URL")
	}
	if raw == "" { // Handle empty strings as unspecified.
		return nil
	}
	url, err := url.Parse(raw)
	if err != nil {
		return errors.Wrap(err, "parse URL")
	}

	if !url.IsAbs() || url.Opaque != "" { // Do not allow opaque URLs.
		return errors.Errorf("not an absolute URL: %s", raw)
	}

	u.Raw = raw
	u.URL = url
	return nil
}

// String returns the raw original URL string.
func (u *URL) String() string {
	return u.Raw
}

// Seconds is a time.Duration which is represented in configuration files as a
// whole number of seconds.
type Seconds time.Duration

// Or returns the time.Duration represented by s or def if s is zero.
func (s *Seconds) Or(def time.Duration) time.Duration {
	if *s > 0 {
		return time.Duration(*s)
	}
	return def
}

// UnmarshalJSON unmarshals Seconds from a JSON number.
func (s *Seconds) UnmarshalJSON(data []byte) error {
	if noop(data) {
		return nil
	}

	var duration time.Duration
	if err := json.Unmarshal(data, &duration); err != nil {
		return errors.Wrap(err, "unmarshal seconds")
	}
	*s = Seconds(duration * time.Second)
	return nil
}

// String returns the duration formatted as a string.
func (s Seconds) String() string {
	return time.Duration(s).String()
}

// Date is a time.Time which is represented as ISO date (yyyy-mm-dd) in JSON
// representations.
type Date struct {
	time.Time
}

// ParseDate parses Date from a string value.
func ParseDate(s string) (Date, error) {
	if s == "" { // Handle empty strings as unspecified.
		return Date{}, nil
	}

	t, err := time.ParseInLocation(dateLayout, s, time.Local)
	if err != nil {
		return Date{}, errors.Wrap(err, "parse date")
	}

	return Date{Time: t}, nil
}

// AtEndOfDay if date represents zero time, returns it's value, otherwise returns the date value
// at time '23:59:59' in local TZ.
func (d Date) AtEndOfDay() Date {
	if d.IsZero() {
		return d
	}
	return d.At(23, 59, 59)
}

// At returns a new date value with the same date, but time set to 'hh:mm:ss' in local TZ.
func (d Date) At(hh, mm, ss int) Date {
	return Date{Time: time.Date(d.Local().Year(), d.Local().Month(), d.Local().Day(), hh, mm, ss, 00, time.Local)}
}

// Equal checks if the d is equal with a given Date or time.Time.
func (d Date) Equal(other interface{}) bool {
	switch t := other.(type) {
	case Date:
		return d.At(00, 00, 00).Time.Equal(t.At(00, 00, 00).Time)
	case time.Time:
		return d.Time.Equal(t)
	}
	return false
}

// UnmarshalJSON unmarshals Date from a JSON string.
func (d *Date) UnmarshalJSON(data []byte) error {
	if noop(data) {
		return nil
	}

	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return errors.Wrap(err, "unmarshal date")
	}
	var err error
	*d, err = ParseDate(raw)
	return err
}

// MarshalJSON marshals the Date to string.
func (d Date) MarshalJSON() ([]byte, error) {
	return wrapMarshalError("marshal date")(json.Marshal(
		formatTime(d.Time, time.Local, dateLayout)))
}

// Time is a time.Time which is represented as clock (HH:MM) in JSON representations.
type Time struct {
	time.Time
}

// UnmarshalJSON unmarshals Time from a JSON string.
func (d *Time) UnmarshalJSON(data []byte) error {
	if noop(data) {
		return nil
	}

	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return errors.Wrap(err, "unmarshal time")
	}
	var err error
	*d, err = ParseTime(raw)
	return err
}

// ParseTime parses Time from a string value.
func ParseTime(s string) (Time, error) {
	if s == "" { // Handle empty strings as unspecified.
		return Time{}, nil
	}

	t, err := time.Parse(timeLayout, s)
	if err != nil {
		return Time{}, errors.Wrap(err, "parse time")
	}

	return Time{Time: t}, nil
}

// MarshalJSON marshals the Time to string.
func (d Time) MarshalJSON() ([]byte, error) {
	return wrapMarshalError("marshal time")(json.Marshal(
		formatTime(d.Time, time.UTC, timeLayout)))
}

// DateTime is a time.Time which is represented as ISO datetime
// (yyyy-mm-ddTHH:MM:SSZ07:00) in JSON representations. It is needed as a
// wrapper around time.Time to marshal the zero time as an empty string.
type DateTime struct {
	time.Time
}

// ParseDateTime parses DateTime from a string value.
func ParseDateTime(s string) (DateTime, error) {
	if s == "" { // Handle empty strings as unspecified.
		return DateTime{}, nil
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return DateTime{}, errors.Wrap(err, "parse datetime")
	}

	return DateTime{Time: t}, nil
}

// UnmarshalJSON unmarshals DateTime from a JSON string.
func (d *DateTime) UnmarshalJSON(data []byte) error {
	if noop(data) {
		return nil
	}

	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return errors.Wrap(err, "unmarshal datetime")
	}
	var err error
	*d, err = ParseDateTime(raw)
	return err
}

// MarshalJSON marshals the DateTime to string.
func (d DateTime) MarshalJSON() ([]byte, error) {
	return wrapMarshalError("marshal datetime")(json.Marshal(
		formatTime(d.Time, d.Location(), time.RFC3339)))
}

// Duration is a time.Time which is represented as ISO date (yyyy-mm-dd) in JSON
// representations.
type Duration struct {
	time.Duration
}

// ParseDuration parses Duration from a string value.
func ParseDuration(s string) (Duration, error) {
	if s == "" { // Handle empty strings as unspecified.
		return Duration{}, nil
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return Duration{}, errors.Wrap(err, "parse duration")
	}
	return Duration{Duration: d}, nil
}

// Or returns the time.Duration represented by d or def if s is zero.
func (d *Duration) Or(def time.Duration) time.Duration {
	if d.Duration > 0 {
		return d.Duration
	}
	return def
}

// UnmarshalJSON unmarshals Duration from a JSON string.
func (d *Duration) UnmarshalJSON(data []byte) error {
	if noop(data) {
		return nil
	}

	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return errors.Wrap(err, "unmarshal duration")
	}
	var err error
	*d, err = ParseDuration(raw)
	return err
}

// MarshalJSON marshals the Duration to string.
func (d Duration) MarshalJSON() ([]byte, error) {
	return wrapMarshalError("marshal duration")(json.Marshal(func() string {
		if d.Duration == 0 {
			return ""
		}
		return d.String()
	}()))
}

// formatTime formats the given time translated to location with the given
// format, with special treatment of returning "" for zero time.
func formatTime(t time.Time, location *time.Location, format string) string {
	if t.IsZero() {
		return ""
	}
	return t.In(location).Format(format)
}

// wrapMarshalError is helper for wrapping the error returned from json.Marshal, if present, with
// the given message.
func wrapMarshalError(s string) func([]byte, error) ([]byte, error) {
	return func(bytes []byte, err error) ([]byte, error) {
		if err != nil {
			return nil, errors.Wrap(err, s)
		}
		return bytes, nil
	}
}

func noop(data []byte) bool {
	// Ignore null, required by interface definition.
	return string(data) == "null"
}
