package log

import (
	"encoding/json"
	"io"
	"os"
	"sync"
)

// syncOutput groups a mutual exclusion lock, an io.Writer, and a JSON-encoder
// which outputs to the Writer into a synchronous log output. The mutex must be
// manually locked when Write is called - either directly or via the encoder.
// This is not performed automatically in order to allow multiple calles to
// Write without releasing the lock.
type syncOutput struct {
	sync.Mutex
	io.Writer
	*json.Encoder
}

func newSyncOutput(w io.Writer) *syncOutput {
	return &syncOutput{Writer: w, Encoder: json.NewEncoder(w)}
}

// syncStderr is a syncOutput that outputs to os.Stderr. It must be shared
// between all loggers writing to stderr to avoid mixing their output.
//
// nolint: gochecknoglobals, runtime constant value.
var syncStderr = newSyncOutput(os.Stderr)
