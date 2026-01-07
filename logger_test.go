package sloghuman_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	logger "github.com/tmstorm/slog-human"

	"github.com/stretchr/testify/assert"
)

type request struct {
	Type      string
	Status    string
	Method    string
	Path      string
	Remote    string
	RequestID string
	Bytes     string
	Duration  string
	Message   string
	Attrs     []any
}

func TestNewDefaultLogger_CallsUnderlying(t *testing.T) {
	l := logger.NewDefaultLogger()
	assert.NotNil(t, l)
}

func TestNewDefaultLoggerTo_Output(t *testing.T) {
	var buf bytes.Buffer
	a := assert.New(t)

	l := logger.NewDefaultLoggerTo(&buf)

	l.Info("default logger test", slog.String("foo", "bar"))
	l.Debug("should not appear")

	out := buf.String()
	a.Contains(out, "default logger test")
	a.Contains(out, "foo=bar")
	a.NotContains(out, "should not appear")
}

func TestNewDefaultLoggerTo_NullWriter(t *testing.T) {
	a := assert.New(t)

	out := captureStdout(func() {
		l := logger.NewDefaultLoggerTo(nil)
		l.Info("multi-handler test", slog.String("foo", "bar"))
	})

	a.Contains(out, "handler has nil Writer")
	a.Contains(out, "multi-handler test")
	a.Contains(out, "foo=bar")
}

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	var buf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		io.Copy(&buf, r)
	}()

	f()
	w.Close()
	wg.Wait()
	os.Stdout = old
	return buf.String()
}

func TestNewLoggerMultiHandler_TextLogger_HTTPRequestFormat(t *testing.T) {
	var buf bytes.Buffer
	a := assert.New(t)
	r := request{
		Type:      "http_request",
		Status:    "200",
		Method:    "GET",
		Path:      "/health",
		Remote:    "localhost",
		RequestID: "hostname",
		Bytes:     "24",
		Duration:  "25µs",
		Attrs: []any{
			slog.String("foo", "bar"),
		},
	}

	l := logger.NewLoggerMultiHandler(logger.Handler{
		Type:   logger.LoggerTypeText,
		Writer: &buf,
		Opts: &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		},
	})

	l.With(
		slog.String("status", r.Status),
		slog.String("method", r.Method),
		slog.String("path", r.Path),
		slog.String("remote", r.Remote),
		slog.String("request_id", r.RequestID),
		slog.String("bytes", r.Bytes),
		slog.String("duration", r.Duration),
		slog.String("log_type", r.Type),
	).Info(r.Type, r.Attrs...)

	out := buf.String()

	a.Contains(out, r.Type)
	a.Contains(out, r.Status)
	a.Contains(out, r.Method)
	a.Contains(out, r.Path)
	a.Contains(out, r.Remote)
	a.Contains(out, r.RequestID)
	a.Contains(out, r.Bytes)
	a.Contains(out, r.Duration)
	for _, v := range r.Attrs {
		attr := v.(slog.Attr)
		formatted := fmt.Sprintf("%s=%v", attr.Key, attr.Value)
		a.Contains(out, formatted)
	}
}

func TestNewLoggerMultiHandler_NullOpts(t *testing.T) {
	var buf bytes.Buffer
	a := assert.New(t)

	l := logger.NewLoggerMultiHandler(
		[]logger.Handler{
			{
				Type:   logger.LoggerTypeText,
				Writer: &buf,
			},
			{
				Type:   logger.LoggerTypeJSON,
				Writer: &buf,
			},
		}...)

	l.Info("multi-handler test", slog.String("foo", "bar"))

	out := buf.String()

	// stdout
	a.Contains(out, "multi-handler test")
	a.Contains(out, "foo=bar")

	// json
	a.Contains(out, `"msg":"multi-handler test"`)
	a.Contains(out, `"foo":"bar"`)
}

func TestNewLoggerMultiHandler_HTTPWriter(t *testing.T) {
	a := assert.New(t)
	var buf bytes.Buffer
	var body []byte

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.Equal(http.MethodPost, r.Method)

		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer s.Close()

	w := logger.NewHTTPWriter(s.URL)

	l := logger.NewLoggerMultiHandler([]logger.Handler{
		{
			Type:   logger.LoggerTypeJSON,
			Writer: w,
		},
		{
			Type:   logger.LoggerTypeJSON,
			Writer: &buf,
		},
	}...)

	l.Info("", slog.String("log_type", "http_request"))
	bufOut := buf.String()

	a.Equal(bufOut, string(body))
}

func TestNewLoggerMultiHandler_MultipleHandlers(t *testing.T) {
	var buf bytes.Buffer
	a := assert.New(t)

	l := logger.NewLoggerMultiHandler(
		[]logger.Handler{
			{
				Type:   logger.LoggerTypeText,
				Writer: &buf,
				Opts: &slog.HandlerOptions{
					Level:     slog.LevelDebug,
					AddSource: true,
				},
			},
			{
				Type:   logger.LoggerTypeJSON,
				Writer: &buf,
				Opts: &slog.HandlerOptions{
					Level:     slog.LevelDebug,
					AddSource: true,
				},
			},
		}...)

	l.Info("multi-handler test", slog.String("foo", "bar"))

	out := buf.String()

	// stdout
	a.Contains(out, "multi-handler test")
	a.Contains(out, "foo=bar")

	// json
	a.Contains(out, `"msg":"multi-handler test"`)
	a.Contains(out, `"foo":"bar"`)
}

func TestNewLoggerMultiHandler_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	a := assert.New(t)

	l := logger.NewLoggerMultiHandler(
		[]logger.Handler{
			{
				Type:   logger.LoggerTypeText,
				Writer: &buf,
				Opts: &slog.HandlerOptions{
					Level:     slog.LevelInfo,
					AddSource: true,
				},
			},
			{
				Type:   logger.LoggerTypeJSON,
				Writer: &buf,
				Opts: &slog.HandlerOptions{
					Level:     slog.LevelInfo,
					AddSource: true,
				},
			},
		}...)

	r := request{
		Type:      "http_request",
		Status:    "200",
		Method:    "GET",
		Path:      "/health",
		Remote:    "localhost",
		RequestID: "hostname",
		Bytes:     "24",
		Duration:  "25µs",
		Attrs: []any{
			slog.String("foo", "bar"),
		},
	}

	l.With(
		slog.String("log_type", r.Type),
		slog.String("status", r.Status),
		slog.String("method", r.Method),
		slog.String("path", r.Path),
		slog.String("remote", r.Remote),
		slog.String("request_id", r.RequestID),
		slog.String("bytes", r.Bytes),
		slog.String("duration", r.Duration),
	).Info(r.Type, r.Attrs...)

	out := buf.String()

	// stdout
	a.Contains(out, r.Type)
	a.Contains(out, r.Status)
	a.Contains(out, r.Method)
	a.Contains(out, r.Path)
	a.Contains(out, r.Remote)
	a.Contains(out, r.RequestID)
	a.Contains(out, r.Bytes)
	a.Contains(out, r.Duration)
	for _, v := range r.Attrs {
		attr := v.(slog.Attr)
		formatted := fmt.Sprintf("%s=%v", attr.Key, attr.Value)
		a.Contains(out, formatted)
	}

	// json
	a.Contains(out, fmt.Sprintf(`"log_type":"%s"`, r.Type))
	a.Contains(out, fmt.Sprintf(`"status":"%s"`, r.Status))
	a.Contains(out, fmt.Sprintf(`"method":"%s"`, r.Method))
	a.Contains(out, fmt.Sprintf(`"path":"%s"`, r.Path))
	a.Contains(out, fmt.Sprintf(`"bytes":"%s"`, r.Bytes))
	a.Contains(out, fmt.Sprintf(`"duration":"%s"`, r.Duration))
	for _, v := range r.Attrs {
		attr := v.(slog.Attr)
		formatted := fmt.Sprintf(`"%s":"%v"`, attr.Key, attr.Value)
		a.Contains(out, formatted)
	}
}

func TestNewLoggerMultiHandler_WithGroup(t *testing.T) {
	var buf bytes.Buffer
	a := assert.New(t)

	l := logger.NewLoggerMultiHandler(
		[]logger.Handler{
			{
				Type:   logger.LoggerTypeText,
				Writer: &buf,
				Opts: &slog.HandlerOptions{
					Level:     slog.LevelInfo,
					AddSource: true,
				},
			},
			{
				Type:   logger.LoggerTypeJSON,
				Writer: &buf,
				Opts: &slog.HandlerOptions{
					Level:     slog.LevelInfo,
					AddSource: true,
				},
			},
		}...)

	l.WithGroup("user").With(
		slog.String("id", "123"),
		slog.String("location", "home"),
	).Info("")

	out := buf.String()

	// stdout
	a.Contains(out, "user.id=123")
	a.Contains(out, "user.location=home")

	// json
	a.Contains(out, `"user":{"id":"123","location":"home"}`)
}

func TestSetLoggerAdapter_SetSlogDefault(t *testing.T) {
	var buf bytes.Buffer
	a := assert.New(t)
	l := logger.NewDefaultLoggerTo(&buf)

	slog.SetDefault(l)

	slog.Info("from slog adapter", slog.String("foo", "bar"))

	out := buf.String()
	a.Contains(out, "from slog adapter")
	a.Contains(out, "foo=bar")
}

func TestLoggerInterface(t *testing.T) {
	var buf bytes.Buffer
	a := assert.New(t)

	l := logger.NewDefaultLoggerTo(&buf)

	l.Info("info test", "foo", "bar")
	l.Warn("warn test", "code", 123)
	l.Error("error test")

	out := buf.String()

	a.Contains(out, "info test")
	a.Contains(out, "foo=bar")
	a.Contains(out, "warn test")
	a.Contains(out, "code=123")
	a.Contains(out, "error test")
}

func TestLoggerInterface_WithContext(t *testing.T) {
	var buf bytes.Buffer
	a := assert.New(t)

	l := logger.NewDefaultLoggerTo(&buf)

	ctx := context.Background()
	l.InfoContext(ctx, "info context test", "foo", "bar")
	l.WarnContext(ctx, "warn context test", "code", 123)
	l.ErrorContext(ctx, "error context test")

	out := buf.String()

	a.Contains(out, "info context test")
	a.Contains(out, "foo=bar")
	a.Contains(out, "warn context test")
	a.Contains(out, "code=123")
	a.Contains(out, "error context test")
}

func TestLoggerInterface_DebugLevels(t *testing.T) {
	var buf bytes.Buffer
	a := assert.New(t)

	l := logger.NewLoggerMultiHandler(logger.Handler{
		Type:   logger.LoggerTypeText,
		Writer: &buf,
		Opts: &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		},
	})

	ctx := context.Background()
	l.Debug("debug test", "foo", "bar")
	l.DebugContext(ctx, "debug context test")

	out := buf.String()

	a.Contains(out, "debug test")
	a.Contains(out, "foo=bar")
	a.Contains(out, "debug context test")
}

func TestColorize_DefaultColors_AppliedToAllCases(t *testing.T) {
	var buf bytes.Buffer
	a := assert.New(t)
	timeNow := time.Now()
	l := logger.NewLoggerMultiHandler(logger.Handler{
		Type:   logger.LoggerTypeText,
		Writer: &buf,
		Opts: &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		},
	})

	c := logger.Colors

	// One log per method case
	l.Info("", slog.String("method", "GET"), slog.String("log_type", "http_request"))
	l.Info("", slog.String("method", "POST"), slog.String("log_type", "http_request"))
	l.Info("", slog.String("method", "PUT"), slog.String("log_type", "http_request"))
	l.Info("", slog.String("method", "DELETE"), slog.String("log_type", "http_request"))
	l.Info("", slog.String("method", "OPTIONS"), slog.String("log_type", "http_request"))
	l.Info("", slog.String("method", "PATCH"), slog.String("log_type", "http_request"))
	l.Info("", slog.String("method", "CONNECT"), slog.String("log_type", "http_request"))
	l.Info("", slog.String("method", "TRACE"), slog.String("log_type", "http_request"))
	l.Info("", slog.String("method", "HEAD"), slog.String("log_type", "http_request"))

	// One log per status case
	l.Info("", slog.String("status", "200"), slog.String("log_type", "http_request"))
	l.Info("", slog.String("status", "301"), slog.String("log_type", "http_request"))
	l.Info("", slog.String("status", "404"), slog.String("log_type", "http_request"))
	l.Info("", slog.String("status", "500"), slog.String("log_type", "http_request"))
	l.Info("", slog.String("status", "100"), slog.String("log_type", "http_request"))

	// One log per level
	l.Debug("DEBUG")
	l.Info("INFO")
	l.Warn("WARN")
	l.Error("ERROR")

	// Other values
	l.Info("", slog.String("request_id", "abc123"), slog.String("log_type", "http_request"))
	l.Info("", slog.String("path", "/health"), slog.String("log_type", "http_request"))
	l.Info("", slog.String("remote", "remote"), slog.String("log_type", "http_request"))

	timeSince := time.Since(timeNow)
	l.Info("", slog.String("duration", timeSince.String()), slog.String("log_type", "http_request"))
	l.Info("", slog.String("bytes", "24"), slog.String("log_type", "http_request"))

	// Additional attrs

	out := buf.String()

	a.Contains(out, c.MethodGET+"GET")
	a.Contains(out, c.MethodPOST+"POST")
	a.Contains(out, c.MethodPUT+"PUT")
	a.Contains(out, c.MethodDELETE+"DELETE")
	a.Contains(out, c.MethodPATCH+"PATCH")
	a.Contains(out, c.MethodOPTIONS+"OPTIONS")
	a.Contains(out, c.MethodCONNECT+"CONNECT")
	a.Contains(out, c.MethodTRACE+"TRACE")
	a.Contains(out, c.MethodHEAD+"HEAD")

	a.Contains(out, c.Status2xx+"200")
	a.Contains(out, c.Status3xx+"301")
	a.Contains(out, c.Status4xx+"404")
	a.Contains(out, c.Status5xx+"500")
	a.Contains(out, "100") // not colored

	a.Contains(out, c.LevelDEBUG+"DEBUG")
	a.Contains(out, c.LevelINFO+"INFO")
	a.Contains(out, c.LevelWARN+"WARN")
	a.Contains(out, c.LevelERROR+"ERROR")

	a.Contains(out, c.RequestID+"abc123")
	a.Contains(out, c.Path+"/health")
	a.Contains(out, "remote")
	a.Contains(out, timeSince.String())
	a.Contains(out, "24B")

	// Check for dangling color codes
	a.NotContains(out, "\033[0m\033")
}

func TestNoColor(t *testing.T) {
	var buf bytes.Buffer
	os.Setenv("NO_COLOR", "")
	a := assert.New(t)

	l := logger.NewLoggerMultiHandler(logger.Handler{
		Type:   logger.LoggerTypeText,
		Writer: &buf,
		Opts: &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		},
	})

	c := logger.Colors

	l.Info("", slog.String("path", "/health"), slog.String("log_type", "http_request"))

	out := buf.String()
	a.NotContains(out, c.Path+"/health"+c.Reset)
	a.Contains(out, "/health")
}
