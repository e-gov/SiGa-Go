package log

import "os"

// systemd speculates if the application is started under systemd. Imitating
// sd_booted is not enough because we want to know if the cmd was actually
// started by the service manager. Check for INVOCATION_ID instead.
//
// nolint: gochecknoglobals, runtime constant value.
var systemd = len(os.Getenv("INVOCATION_ID")) > 0

// nolint: gochecknoglobals, constant map.
var systemdPrefix = map[Level]string{
	LevelError: "<3>",
	LevelInfo:  "<6>",
	LevelDebug: "<7>",
}
