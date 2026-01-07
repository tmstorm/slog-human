// Package sloghuman
package sloghuman

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type (
	// TextHandler is used to create and implement the slog-human TextHandler
	TextHandler struct {
		mx        sync.Mutex
		out       io.Writer
		level     slog.Level
		addSource bool
		attrs     []slog.Attr
		group     string
		noColor   bool
	}

	// Handler is use to create and acccess a logger handler(s)
	Handler struct {
		Type   LoggerType
		Writer io.Writer
		Opts   *slog.HandlerOptions
	}

	// multiHandler is used by slog-human to acccess all handlers created
	multiHandler struct {
		handlers []slog.Handler
	}

	// LoggerType is the type of handler to be created by slog-human
	LoggerType int
)

// defaultTimeFormat is the default time format.
const defaultTimeFormat = " 2006/01/02 - 15:05:05"

// TextTimeFormat can be used to change the time format when using
// the text handler if none is set defaultTimeFormat is used.
// See https://go.dev/src/time/format.go for setting the time format.
var TextTimeFormat = defaultTimeFormat

// Enums used by slog-human to determine the log handler type
const (
	LoggerTypeText LoggerType = iota
	LoggerTypeJSON
)

func (t LoggerType) String() string {
	return [...]string{"Text", "JSON"}[t]
}

// NewDefaultLogger creates and returns a new text logger with a os.Stdout writer.
func NewDefaultLogger() *slog.Logger {
	return NewDefaultLoggerTo(os.Stdout)
}

// NewDefaultLoggerTo creates and returns a new text logger with the provided writer.
func NewDefaultLoggerTo(w io.Writer) *slog.Logger {
	return NewLoggerMultiHandler(Handler{
		Type:   LoggerTypeText,
		Writer: w,
		Opts: &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		},
	})
}

// NewLoggerMultiHandler returns a new logger with the provided handler(s).
// NOTE: If a handler has a nil writer it will be defaulted to os.Stdout with a warning to os.Stderr
func NewLoggerMultiHandler(handlers ...Handler) *slog.Logger {
	var slogHandlers []slog.Handler

	for _, t := range handlers {
		if t.Writer == nil {
			fmt.Fprintf(os.Stderr, "[slog-human] warning: %s handler has nil Writer - defaulting to os.Stdout\n", t.Type.String())
			t.Writer = os.Stdout
		}

		if t.Opts == nil {
			t.Opts = &slog.HandlerOptions{
				Level:     slog.LevelInfo,
				AddSource: false,
			}
		}

		switch t.Type {
		case LoggerTypeText:
			slogHandlers = append(slogHandlers, newTextHandler(t.Writer, t.Opts))
		case LoggerTypeJSON:
			slogHandlers = append(slogHandlers, slog.NewJSONHandler(t.Writer, t.Opts))
		}
	}

	var handler slog.Handler
	if len(handlers) == 1 {
		handler = slogHandlers[0]
	} else {
		handler = setMultiHandlers(slogHandlers...)
	}

	l := slog.New(handler)
	return l
}

// newTextHandler is the internal helper that creates the text handler used
// by slog-human to print text logs.
func newTextHandler(out io.Writer, opts *slog.HandlerOptions) slog.Handler {
	noColor := false
	if v, ok := os.LookupEnv("NO_COLOR"); ok {
		if strings.ToLower(strings.TrimSpace(v)) != "false" {
			noColor = true
		}
	}

	return &TextHandler{
		out:       out,
		level:     opts.Level.Level(),
		addSource: opts.AddSource,
		noColor:   noColor,
	}
}

// Enabled is the slog-human implementation of slog.Handler interface
func (h *TextHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle is the slog-human implementation of slog.Handler interface.
// The known attrs will all be colored according to the color palette in
// use. See colorize.go for more.
func (h *TextHandler) Handle(_ context.Context, r slog.Record) error {
	h.mx.Lock()
	defer h.mx.Unlock()

	var (
		logType   string
		method    string
		path      string
		status    string
		remote    string
		duration  string
		requestID string
		bytes     string
		fields    []string
		seen      = make(map[string]struct{})
	)

	// Record attrs
	r.Attrs(func(attr slog.Attr) bool {
		switch attr.Key {
		case "log_type":
			logType = attr.Value.String()
		case "method":
			method = attr.Value.String()
		case "path":
			path = attr.Value.String()
		case "status":
			status = attr.Value.String()
		case "duration":
			if d, ok := attr.Value.Any().(time.Duration); ok {
				duration = d.String()
			} else {
				duration = attr.Value.String()
			}
		case "remote":
			remote = attr.Value.String()
		case "request_id":
			requestID = attr.Value.String()
		case "bytes":
			bytes = attr.Value.String()
		default:
			fields = append(fields, fmt.Sprintf("%s=%v", attr.Key, attr.Value))
			seen[attr.Key] = struct{}{}
		}
		return true
	})

	// Handler attrs
	for _, attr := range h.attrs {
		switch attr.Key {
		case "log_type":
			if logType == "" {
				logType = attr.Value.String()
			}
		case "method":
			if method == "" {
				method = attr.Value.String()
			}
		case "path":
			if path == "" {
				path = attr.Value.String()
			}
		case "status":
			if status == "" {
				status = attr.Value.String()
			}
		case "duration":
			if duration == "" {
				if d, ok := attr.Value.Any().(time.Duration); ok {
					duration = d.String()
				} else {
					duration = attr.Value.String()
				}
			}
		case "remote":
			if remote == "" {
				remote = attr.Value.String()
			}
		case "request_id":
			if requestID == "" {
				requestID = attr.Value.String()
			}
		case "bytes":
			if bytes == "" {
				bytes = attr.Value.String()
			}
		default:
			if _, dup := seen[attr.Key]; !dup {
				fields = append(fields, fmt.Sprintf("%s=%v", attr.Key, attr.Value))
			}
		}
	}

	// Colorize
	coloredPath := h.colorize(path, ColorPath)
	coloredStatus := h.colorize(status, ColorStatus)
	coloredRequestID := h.colorize(requestID, ColorRequestID)
	coloredMessage := h.colorize(r.Message, ColorMessage)
	coloredLogType := h.colorize(logType, ColorLogType)

	// Pad method and level to keep lines pretty
	paddedMethod := fmt.Sprintf("%-7s", method)
	coloredMethod := h.colorize(paddedMethod, ColorMethod)
	paddedLevel := fmt.Sprintf("%-5s", r.Level.String())
	coloredLevel := h.colorize(paddedLevel, ColorLevel)

	// Source
	source := ""
	if h.addSource && r.PC != 0 {
		frames := runtime.CallersFrames([]uintptr{r.PC})
		frame, _ := frames.Next()
		if frame.File != "" {
			// Line is padded to hundreds place. After line 999 lines will start to offset
			// from anything under 1000.
			line := h.colorize(fmt.Sprintf("%-3s", strconv.Itoa(frame.Line)), ColorLine)
			source = fmt.Sprintf(" (%s:%s)", filepath.Base(frame.File), line)
		}
	}

	// RequestID prefix
	reqIDPrefix := ""
	if requestID != "" {
		reqIDPrefix = fmt.Sprintf(" [%s]", coloredRequestID)
	}

	// Bytes suffix
	bytesSuffix := bytes
	if bytes != "" {
		bytesSuffix = bytes + "B"
	}

	// Bytes and duration section
	bytesTime := ""
	if bytesSuffix != "" || duration != "" {
		bytesTime = fmt.Sprintf(" [%5s %10s] ", bytesSuffix, duration)
	}

	// Build line
	ts := r.Time.Format(TextTimeFormat)
	var line string

	// check log type and create log line accordingly
	switch logType {
	case "http_request":
		displayType := h.colorize("HTTP Request", ColorLogType)
		line = fmt.Sprintf("[%s]%s%s:%s | %s | %s %s %s %s%s %s",
			coloredLevel, reqIDPrefix, ts, source, displayType,
			coloredStatus, coloredMethod, coloredPath, remote,
			bytesTime, coloredMessage,
		)
	default:
		typePrefix := ""
		if logType != "" {
			typePrefix = " | " + coloredLogType
		}
		line = fmt.Sprintf("[%s]%s:%s%s | %s", coloredLevel, ts, source, typePrefix, coloredMessage)
	}

	if len(fields) > 0 {
		line += " " + strings.Join(fields, " ")
	}
	line += "\n"

	fmt.Fprint(h.out, line)
	return nil
}

// WithAttrs is the slog-human implementation of slog.Handler interface
func (h *TextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	newH := &TextHandler{
		out:       h.out,
		level:     h.level,
		addSource: h.addSource,
		noColor:   h.noColor,
		group:     h.group,
		mx:        sync.Mutex{},
	}
	if h.group != "" {
		prefixed := make([]slog.Attr, len(attrs))
		for i, a := range attrs {
			prefixed[i] = slog.Attr{
				Key:   h.group + "." + a.Key,
				Value: a.Value,
			}
		}
		newH.attrs = append(h.attrs, prefixed...)
	} else {
		newH.attrs = append(h.attrs, attrs...)
	}

	return newH
}

// WithGroup is the slog-human implementation of slog.Handler interface
func (h *TextHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	newH := &TextHandler{
		out:       h.out,
		level:     h.level,
		addSource: h.addSource,
		attrs:     h.attrs,
		mx:        sync.Mutex{},
	}

	if h.group == "" {
		newH.group = name
	} else {
		newH.group = h.group + "." + name
	}

	return newH
}

// setMultiHandlers is an internal function used to create a new multiHandler containing the provided handlers
func setMultiHandlers(handlers ...slog.Handler) slog.Handler {
	return &multiHandler{handlers: handlers}
}

// Enabled is the slog-human multiHandler implementation of slog.Handler interface
func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle is the slog-human multiHandler implementation of slog.Handler interface
func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

// WithAttrs is the slog-human multiHandler implementation of slog.Handler interface
func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: newHandlers}
}

// WithGroup is the slog-human multiHandler implementation of slog.Handler interface
func (m *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: newHandlers}
}
