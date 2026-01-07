package sloghuman

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// HTTPWriter is an io.Writer that sends log lines to a remote HTTP endpoint.
// It Buffers input until a new line is seen, then POSTS the complete JSON line.
// It Supports bit asynchronous and synchronous modes.
type HTTPWriter struct {
	client      *http.Client
	endpoint    string
	contentType string
	buffer      bytes.Buffer
	async       bool
}

// NewAsyncHTTPWriter returns an io.Writer that sends logs asynchronously over HTTP.
//
// This is the recommended mode for production use. Logs are sent in a goroutine, so they
// never block is slow down the program.
func NewAsyncHTTPWriter(endpoint string) io.Writer {
	return &HTTPWriter{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		endpoint:    endpoint,
		contentType: "application/json",
		async:       true,
	}
}

// NewHTTPWriter returns an io.Writer that sends logs synchronously over HTTP.
//
// This mode blocks until the HTTP request completes. Use only for testing
// or when you absolutely need confirmation that logs are sent.
// WARNING: Can introduce performance issues or race conditions.
func NewHTTPWriter(endpoint string) io.Writer {
	return &HTTPWriter{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		endpoint:    endpoint,
		contentType: "application/json",
		async:       false,
	}
}

// Write implements io.Writer. It buffers input until a newline '\n' is detected
// then sends the complete log line over HTTP.
//
// Both slog's built-in JSONHandler and this package's TextHandler append '\n'
// to each record, this ensures one HTTP request per log entry.
func (w *HTTPWriter) Write(p []byte) (int, error) {
	w.buffer.Write(p)

	if len(p) > 0 && p[len(p)-1] == '\n' {
		line := make([]byte, w.buffer.Len())
		copy(line, w.buffer.Bytes())
		w.buffer.Reset()

		if w.async {
			go w.send(line)
		} else {
			w.send(line)
		}
	}

	return len(p), nil
}

// send performs the actual HTTP POST of a log line.
// Errors and non-2xx responses are reported to os.Stderr but do not return errors.
// Logging should never fail the program.
func (w *HTTPWriter) send(payload []byte) {
	req, err := http.NewRequest("POST", w.endpoint, bytes.NewReader(payload))
	if err != nil {
		fmt.Fprintf(os.Stderr, "[slog-human] http writer: failed to create request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", w.contentType)

	resp, err := w.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[slog-human] http writer: failed to send log: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "[slog-human] http writer: bad status %d sending log\n", resp.StatusCode)
	}
}
