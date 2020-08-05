package log

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/pkg/errors"
)

// Mock time used for tests.
const nowstr = "2019-06-19T18:01:52.88689868Z"

var now = time.Date(2019, time.June, 19, 18, 01, 52, 886898680, time.UTC)

// runLogTests sets up a test logger, calls testFn, runs the returned table of
// test cases, and compares the logged output to expected output.
func runLogTests(t *testing.T, testFn func(*Logger) []logTest) {
	t.Helper()

	var buf bytes.Buffer
	logger := &Logger{
		ltype: Type("TEST"),
		out:   newSyncOutput(&buf),
		now:   func() time.Time { return now },
		max:   MaxParameterLength,
	}

	for _, test := range testFn(logger) {
		// Run these steps outside of t.Run, so that the reported stack
		// trace belongs to this goroutine, not the sub-test.
		buf.Reset()
		test.msg.Skip(1).Log(test.ctx, test.tag)
		logged := buf.String()
		logged = logged[:len(logged)-1] // Trim newline.

		t.Run(test.tag, func(t *testing.T) {
			expected := test.output
			ok := logged == expected
			if len(test.pattern) > 0 {
				expected = test.pattern
				ok, _ = regexp.MatchString(expected, logged)
			}
			if !ok {
				t.Errorf("unexpected log message,\nlogged:   %s\nexpected: %s",
					logged, expected)
			}
		})
	}
}

type logTest struct {
	msg *Message        // Preconstructed message to log.
	ctx context.Context // Context to log the message with.
	tag string          // Tag to log the message with.

	output  string // Logged message must equal output (unless pattern is set).
	pattern string // Logged message must match pattern (if set).
}

func TestMessageLog_LevelAndTag_LogsLevelAndTag(t *testing.T) {
	runLogTests(t, func(logger *Logger) []logTest {
		expected := func(level, tag string) string {
			return fmt.Sprintf(`{`+
				`"timestamp":"`+nowstr+`",`+
				`"type":"TEST",`+
				`"level":"%s",`+
				`"tag":"stash.ria.ee/vis3/vis3-common/pkg/log.TestMessageLog_LevelAndTag_LogsLevelAndTag:%s"`+
				`}`, level, tag)
		}
		return []logTest{
			{logger.Debug(), nil, "debug", expected("DEBUG", "debug"), ""},
			{logger.Info(), nil, "info", expected("INFO", "info"), ""},
			{logger.Error(), nil, "error", expected("ERROR", "error"), ""},
		}
	})
}

func TestMessageLog_WithContext_LogsIdentifiers(t *testing.T) {
	runLogTests(t, func(logger *Logger) []logTest {
		ctx := WithRequestID(context.Background(), "request-identifier")
		ctx = WithSessionID(ctx, "session-identifier")

		const expected = `{` +
			`"timestamp":"` + nowstr + `",` +
			`"type":"TEST",` +
			`"level":"INFO",` +
			`"tag":"stash.ria.ee/vis3/vis3-common/pkg/log.TestMessageLog_WithContext_LogsIdentifiers:ctx",` +
			`"request":"request-identifier",` +
			`"session":"session-identifier"` +
			`}`
		return []logTest{
			{logger.Info(), ctx, "ctx", expected, ""},
		}
	})
}

func TestMessageLog_WithParameters_LogsParameterValues(t *testing.T) {
	runLogTests(t, func(logger *Logger) []logTest {
		const fname = "stash.ria.ee/vis3/vis3-common/pkg/log.TestMessageLog_WithParameters_LogsParameterValues"
		const prefix = `{` +
			`"timestamp":"` + nowstr + `",` +
			`"type":"TEST",` +
			`"level":"INFO",` +
			`"tag":"` + fname
		expected := func(tag, params string) string {
			return fmt.Sprintf(prefix+`:%s","params":%s}`, tag, params)
		}

		shortLogger := *logger
		shortLogger.max = 5

		return []logTest{
			{
				msg:    logger.Info().WithString("sprint", "foo", 123, true),
				tag:    "string",
				output: expected("string", `{"sprint":"foo 123 true"}`),
			},
			{
				msg:    logger.Info().WithStringf("sprintf", "foo %d %t", 123, true),
				tag:    "stringf",
				output: expected("stringf", `{"sprintf":"foo 123 true"}`),
			},
			{
				msg:    logger.Info().WithJSON("json", map[string]int{"nested": 123}),
				tag:    "json",
				output: expected("json", `{"json":"{\"nested\":123}"}`),
			},
			{
				msg:    logger.Info().WithError(fmt.Errorf("something failed")),
				tag:    "error",
				output: expected("error", `{"error":"something failed"}`),
			},
			{
				msg: logger.Info().WithError(
					errors.WithMessage(errors.New("something failed"), "cause")),
				tag: "stacktrace",
				pattern: `^` + regexp.QuoteMeta(prefix+`:stacktrace",`+
					`"params":{`+
					`"error":"cause: something failed",`+
					`"stacktrace":"`+fname+`.func1\n\t`) +
					`.+/pkg/log/log_test.go:\d+\\n.*\}\}$`,
			},
			{
				msg:    shortLogger.Info().WithString("long", "truncated"),
				tag:    "length",
				output: expected("length", `{"long":"trunc","long.length":9}`),
			},
		}
	})
}

func TestMessageLog_WithSystemd_LogsLevelPrefix(t *testing.T) {
	systemd = true
	runLogTests(t, func(logger *Logger) []logTest {
		expected := func(prefix, level, tag string) string {
			return fmt.Sprintf(`%s{`+
				`"timestamp":"`+nowstr+`",`+
				`"type":"TEST",`+
				`"level":"%s",`+
				`"tag":"stash.ria.ee/vis3/vis3-common/pkg/log.TestMessageLog_WithSystemd_LogsLevelPrefix:%s"`+
				`}`, prefix, level, tag)
		}
		return []logTest{
			{logger.Debug(), nil, "debug", expected("<7>", "DEBUG", "debug"), ""},
			{logger.Info(), nil, "info", expected("<6>", "INFO", "info"), ""},
			{logger.Error(), nil, "error", expected("<3>", "ERROR", "error"), ""},
		}
	})
	systemd = false
}
