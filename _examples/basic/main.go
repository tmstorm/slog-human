package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	logger "github.com/tmstorm/slog-human"
)

func main() {
	l := logger.NewLoggerMultiHandler(logger.Handler{
		Type:   logger.LoggerTypeText,
		Writer: os.Stdout,
		Opts: &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		},
	})
	t1 := time.Now()

	fmt.Println("\n Dracula")
	log(l, t1)

	fmt.Println("\n GruvBox Dark")
	logger.Colors = logger.GruvboxDark
	log(l, t1)

	fmt.Println("\n One Dark")
	logger.Colors = logger.OneDark
	log(l, t1)

	fmt.Println("\n Nord")
	logger.Colors = logger.Nord
	log(l, t1)

	fmt.Println("\n Solarized Dark")
	logger.Colors = logger.SolarizedDark
	log(l, t1)

	// Disable colors using NO_COLOR environment variable
	os.Setenv("NO_COLOR", "1")
	fmt.Println("\n NO_COLOR")
	noColorLog := logger.NewDefaultLogger()

	// Set slog-simple as default slog logger
	slog.SetDefault(noColorLog)
	slog.Info("slog-simple without colors", slog.String("log_type", "Internal"), slog.String("NO_COLOR", os.Getenv("NO_COLOR")))
}

func log(l *slog.Logger, startTime time.Time) {
	l.Debug("",
		slog.String("log_type", "http_request"),
		slog.String("method", http.MethodGet),
		slog.String("path", "/test"),
		slog.Int("status", http.StatusOK),
		slog.Int("bytes", 120),
		slog.Duration("duration", time.Now().Sub(startTime)),
		slog.String("request_id", "hostname"),
	)
	l.Info("",
		slog.String("log_type", "http_request"),
		slog.String("method", http.MethodPost),
		slog.String("path", "/test"),
		slog.Int("status", http.StatusTemporaryRedirect),
		slog.Int("bytes", 120),
		slog.Duration("duration", time.Now().Sub(startTime)),
		slog.String("request_id", "hostname"),
	)
	l.Warn("",
		slog.String("log_type", "http_request"),
		slog.String("method", http.MethodPut),
		slog.String("path", "/test"),
		slog.Int("status", http.StatusNotFound),
		slog.Int("bytes", 120),
		slog.Duration("duration", time.Now().Sub(startTime)),
		slog.String("request_id", "hostname"),
	)
	l.Error("",
		slog.String("log_type", "http_request"),
		slog.String("method", http.MethodDelete),
		slog.String("path", "/test"),
		slog.Int("status", http.StatusInternalServerError),
		slog.Int("bytes", 120),
		slog.Duration("duration", time.Now().Sub(startTime)),
		slog.String("request_id", "hostname"),
	)
	l.Info("",
		slog.String("log_type", "http_request"),
		slog.String("method", http.MethodPatch),
		slog.String("path", "/test"),
		slog.Int("status", http.StatusTemporaryRedirect),
		slog.Int("bytes", 120),
		slog.Duration("duration", time.Now().Sub(startTime)),
		slog.String("request_id", "hostname"),
	)
	l.Info("",
		slog.String("log_type", "http_request"),
		slog.String("method", http.MethodOptions),
		slog.String("path", "/test"),
		slog.Int("status", http.StatusTemporaryRedirect),
		slog.Int("bytes", 120),
		slog.Duration("duration", time.Now().Sub(startTime)),
		slog.String("request_id", "hostname"),
	)
	l.Info("",
		slog.String("log_type", "http_request"),
		slog.String("method", http.MethodConnect),
		slog.String("path", "/test"),
		slog.Int("status", http.StatusTemporaryRedirect),
		slog.Int("bytes", 120),
		slog.Duration("duration", time.Now().Sub(startTime)),
		slog.String("request_id", "hostname"),
	)
	l.Info("",
		slog.String("log_type", "http_request"),
		slog.String("method", http.MethodTrace),
		slog.String("path", "/test"),
		slog.Int("status", http.StatusTemporaryRedirect),
		slog.Int("bytes", 120),
		slog.Duration("duration", time.Now().Sub(startTime)),
		slog.String("request_id", "hostname"),
	)
	l.Info("",
		slog.String("log_type", "http_request"),
		slog.String("method", http.MethodHead),
		slog.String("path", "/test"),
		slog.Int("status", http.StatusTemporaryRedirect),
		slog.Int("bytes", 120),
		slog.Duration("duration", time.Now().Sub(startTime)),
		slog.String("request_id", "hostname"),
	)
}
