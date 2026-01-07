package main

import (
	"log/slog"
	"net/http"
	"os"

	logger "github.com/tmstorm/slog-human"
	logmiddleware "github.com/tmstorm/slog-human/middleware/chi"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	l := logger.NewLoggerMultiHandler([]logger.Handler{
		{
			Type:   logger.LoggerTypeText,
			Writer: os.Stdout,
		},
		{
			Type:   logger.LoggerTypeJSON,
			Writer: os.Stdout,
		},
	}...)
	slog.SetDefault(l)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Heartbeat("/health"))
	r.Use(logmiddleware.ChiLogger(slog.Default()))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World"))
	})

	slog.Info("Chi starting on port :3000")
	err := http.ListenAndServe(":3000", r)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
