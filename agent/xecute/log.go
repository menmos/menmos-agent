package xecute

import (
	"encoding/json"
	"io"

	"github.com/menmos/menmos-agent/agent/xecute/ring"
)

const BUFFER_LOG_LINES = 512

type logWriter struct {
	out         io.WriteCloser
	lineBuffer  *ring.Buffer[[]byte]
	currentLine []byte
}

func newLogWriter(out io.WriteCloser) *logWriter {
	return &logWriter{
		out:        out,
		lineBuffer: ring.New[[]byte](BUFFER_LOG_LINES),
	}
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		if p[i] == '\n' {
			w.lineBuffer.Write(w.currentLine)
			w.currentLine = []byte{}
		}
		w.currentLine = append(w.currentLine, p[i])
	}
	return w.out.Write(p)
}

func (w *logWriter) Close() error {
	return w.out.Close()
}

func (w *logWriter) GetLastNLines(n int) (lines []interface{}) {
	if n < BUFFER_LOG_LINES {
		n = BUFFER_LOG_LINES
	}

	// FIXME(prod): swap log files after a given amount of time and/or entries.

	for i := 0; i < n; i++ {
		raw := w.lineBuffer.Peek(i)
		if raw == nil {
			// We're at the end of the line buffer.
			break
		}

		// Try reading the log as JSON.
		var entry map[string]interface{}
		if err := json.Unmarshal(raw, &entry); err != nil {
			// Just add the log as string.
			lines = append(lines, string(raw))
		} else {
			lines = append(lines, entry)
		}
	}

	return
}
