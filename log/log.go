/*
Package log is a logging package for JSON logs.

Log output

Logs are printed to stderr instead of a syslog socket or some other specific
mechanism. This keeps it simple and compatible with service managers (e.g.,
Systemd) or container platforms (e.g., Docker) which expect to collect logs
from services via stderr.

This also behaves as expected when an application is run manually in the
foreground: logs are printed to the terminal.

Log format

All logs are formatted as JSON. Minimally the JSON object contains:

  - a timestamp formatted according to RFC 3339 with fractional seconds,
  - a message type indicator (e.g., "SERVER", "SECURITY"),
  - a message level indicator (e.g., "INFO", "ERROR"), and
  - a unique tag specifying the message.

Additionally, if the message was logged during a request:

  - a client request identifier,

and if the request had an active client session:

  - a client session identifier.

Optionally, the JSON object can contain a sub-object with parameters providing
extra details about the message, for example a stack trace. The values of these
parameters can only be strings (see an exception below). Although the request
identifier and session identifier also behave like parameters, they are
promoted to top-level fields because they are included in the vast majority of
log messages.

In addition to regular JSON-escaping, any characters in parameter values with a
unicode code point value of 127 (DEL) or greater will be escaped.

Encoded values are truncated to a maximum length: if this happens, then a
separate "<key>.length" parameter is added, which records the original length
of the value as a number (not a string). If a "<key>.length" parameter is added
manually, then it will be renamed to "<key>_length".

The truncated value can be shorter than the maximum length if it would have
otherwise ended in the middle of an escaped character (e.g., "\ufffd" truncated
to "\uf"). In this case, the entire encoded character is dropped.
*/
package log

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// MaxParameterLength is the maximum length of an encoded parameter value
// before it is truncated.
const MaxParameterLength = 4096

// Type specifies a log message type.
type Type string

const (
	// TypeServer is the default log type used for most messages.
	TypeServer Type = "SERVER"

	// TypeSecurity is the log type used for security-critical messages.
	TypeSecurity Type = "SECURITY"
)

// Level specifies a log message secerity level.
type Level string

// Enumeration of supported logging levels.
const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelError Level = "ERROR"
)

// key is the key type of values stored in contexts by this package.
type key int

const (
	reqIDKey  key = iota // Context key for request identifier.
	sessIDKey            // Context key for session identifier.
)

// WithRequestID sets the client request identifier for messages logged with
// the returned context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, reqIDKey, id)
}

// WithSessionID sets the client session identifier for messages logged with
// the returned context.
func WithSessionID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, sessIDKey, id)
}

// Logger is a logging object that writes log messages of a specific type.
type Logger struct {
	ltype Type
	out   *syncOutput
	now   func() time.Time
	max   int
}

// New returns a new Logger which logs messages of type ltype.
func New(ltype Type) *Logger {
	return &Logger{
		ltype: ltype,
		out:   syncStderr,
		now:   time.Now,
		max:   MaxParameterLength,
	}
}

func (l *Logger) at(level Level) *Message {
	return &Message{
		logger: l,
		msg: message{
			Timestamp: l.now(),
			Type:      l.ltype,
			Level:     level,
		},
	}
}

// Debug returns a new log message object at LevelDebug.
func (l *Logger) Debug() *Message { return l.at(LevelDebug) }

// Info returns a new log message object at LevelInfo.
func (l *Logger) Info() *Message { return l.at(LevelInfo) }

// Error returns a new log message object at LevelError.
func (l *Logger) Error() *Message { return l.at(LevelError) }

// nolint: gochecknoglobals, runtime constant values.
var (
	// Server is a package-level instance of a SERVER type logger.
	Server = New(TypeServer)

	// Security is a package-level instance of a SECURITY type logger.
	Security = New(TypeSecurity)
)

// Debug calls Server.Debug. Package-level convenience method for default logs.
func Debug() *Message { return Server.Debug() }

// Info calls Server.Info. Package-level convenience method for default logs.
func Info() *Message { return Server.Info() }

// Error calls Server.Error. Package-level convenience method for default logs.
func Error() *Message { return Server.Error() }

// Message is a log message object. Message methods can be used to include
// additional parameters in the message before actually logging it using the
// terminal Log method.
type Message struct {
	logger *Logger
	msg    message
	skip   int // Additional frames to skip when determining caller. Non-zero for testing.
}

type message struct {
	Timestamp time.Time `json:"timestamp"`
	Type      Type      `json:"type"`
	Level     Level     `json:"level"`
	Tag       string    `json:"tag"`

	Request string `json:"request,omitempty"`
	Session string `json:"session,omitempty"`

	Params map[string]interface{} `json:"params,omitempty"`
}

func (m *Message) withParam(key, value string) *Message {
	if m.msg.Params == nil {
		m.msg.Params = make(map[string]interface{})
	}
	const length = ".length"
	if strings.HasSuffix(key, length) {
		key = key[:len(key)-len(length)] + "_length"
	}
	encoded, truncated := encode(value, m.logger.max)
	m.msg.Params[key] = encoded
	if truncated {
		m.msg.Params[key+length] = len(value)
	}
	return m
}

// WithString adds a parameter with the specified key to the Message. Arguments
// are formatted with fmt.Sprintln to form the string value, but the trailing
// newline is removed.
func (m *Message) WithString(key string, args ...interface{}) *Message {
	str := fmt.Sprintln(args...)
	return m.withParam(key, str[:len(str)-1])
}

// WithStringf adds a parameter with the specified key to the Message.
// Arguments are formatted with fmt.Sprintf to form the string value.
func (m *Message) WithStringf(key, format string, args ...interface{}) *Message {
	return m.withParam(key, fmt.Sprintf(format, args...))
}

// WithJSON adds a parameter with the specified key to the Message. The value
// is JSON-encoded and the result is used as the string value.
func (m *Message) WithJSON(key string, value interface{}) *Message {
	// Encode into strings.Builder to avoid string([]byte) alloc.
	var buf strings.Builder
	json.NewEncoder(&buf).Encode(value) // nolint: errcheck
	str := buf.String()
	return m.withParam(key, str[:len(str)-1])
}

// WithError adds an "error" parameter to the Message with the string value of
// err.Error().
//
// Additionally, if err implements the interface
//
//     import "github.com/pkg/errors"
//
//     type stackTracer interface {
//             StackTrace() errors.StackTrace
//     }
//
// then a "stacktrace" parameter is added containing the stack trace of err. If
// instead err implements
//
//     type causer interface {
//             Cause() error
//     }
//
// then the process is repeated with the causing error until a stack trace is
// found or the error no longer implements causer.
func (m *Message) WithError(err error) *Message {
	m.withParam("error", err.Error())
	if trace := findtrace(err); trace != nil {
		if str := fmt.Sprintf("%+v", trace); len(str) > 0 {
			m.withParam("stacktrace", str[1:]) // Starts with a newline.
		}
	}
	return m
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type causer interface {
	Cause() error
}

func findtrace(err error) errors.StackTrace {
	switch t := err.(type) {
	case stackTracer:
		return t.StackTrace()
	case causer:
		// Do not use errors.Cause because it can skip a stackTracer.
		return findtrace(t.Cause())
	default:
		return nil
	}
}

// Skip adds an additional number of stack frames to skip when determining the
// function whose name to use as the prefix in Log. Useful in helper functions
// which want to use the name of their caller.
func (m *Message) Skip(skip int) *Message {
	m.skip += skip
	return m
}

// Log logs the message object to output using the provided context and tag.
// The context is checked for optional request and session identifiers. The tag
// is prefixed by the function name that called Log (see Skip).
func (m *Message) Log(ctx context.Context, tag string) {
	var caller string
	if pc, _, _, ok := runtime.Caller(m.skip + 1); ok {
		caller = runtime.FuncForPC(pc).Name()
	}
	m.msg.Tag = caller + ":" + tag

	if ctx != nil {
		if id := ctx.Value(reqIDKey); id != nil {
			m.msg.Request = id.(string)
		}
		if id := ctx.Value(sessIDKey); id != nil {
			m.msg.Session = id.(string)
		}
	}

	m.logger.out.Lock()
	defer m.logger.out.Unlock()
	if systemd {
		// nolint: errcheck, nowhere to report the error.
		io.WriteString(m.logger.out, systemdPrefix[m.msg.Level])
	}
	m.logger.out.Encode(m.msg) // nolint: errcheck, nowhere to report the error.
}
