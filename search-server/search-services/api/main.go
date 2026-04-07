package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"yadro.com/course/api/adapters/aaa"
	"yadro.com/course/api/adapters/rest"
	"yadro.com/course/api/adapters/rest/middleware"
	"yadro.com/course/api/adapters/search"
	"yadro.com/course/api/adapters/update"
	"yadro.com/course/api/adapters/words"
	"yadro.com/course/api/config"
	"yadro.com/course/api/core"
)

const gracefulShutdownTime = time.Second * 3

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.LogLevel)

	log.Info("starting server")
	log.Debug("debug messages are enabled")

	updateClient, err := update.NewClient(cfg.UpdateAddress, log)
	if err != nil {
		log.Error("cannot init update adapter", "error", err)
		return
	}

	defer func() {
		if err := updateClient.Close(); err != nil {
			log.Error("cant close conn in updateClient", "err", err)
		}
	}()

	searchClient, err := search.NewClient(cfg.SearchAddress, log)
	if err != nil {
		log.Error("cannot init search adapter", "error", err)
		return
	}

	defer func() {
		if err := searchClient.Close(); err != nil {
			log.Error("cant close conn in searchClient", "err", err)
		}
	}()

	wordsClient, err := words.NewClient(cfg.WordsAddress, log)
	if err != nil {
		log.Error("cannot init words adapter", "error", err)
		return
	}
	defer func() {
		if err := wordsClient.Close(); err != nil {
			log.Error("close empty conn")
		}
	}()

	verifier, err := aaa.New(log, cfg.TokenTTL)
	if err != nil {
		log.Error("creating verifier error", "err", err)
		return
	}

	mux := http.NewServeMux()
	mux.Handle("POST /api/db/update", middleware.Metrics(middleware.Auth(rest.NewUpdateHandler(log, updateClient), verifier)))
	mux.Handle("GET /api/db/stats", middleware.Metrics(rest.NewUpdateStatsHandler(log, updateClient)))
	mux.Handle("GET /api/db/status", middleware.Metrics(rest.NewUpdateStatusHandler(log, updateClient)))
	mux.Handle("DELETE /api/db", middleware.Metrics(middleware.Auth(rest.NewDropHandler(log, updateClient), verifier)))
	mux.Handle("GET /api/ping", middleware.Metrics(rest.NewPingHandler(log, map[string]core.Pinger{"words": wordsClient, "update": updateClient, "search": searchClient})))
	mux.Handle("GET /api/search", middleware.Metrics(middleware.Concurrency(rest.NewSearchHandler(log, searchClient), cfg.SearchConcurrency)))
	mux.Handle("GET /api/isearch", middleware.Metrics(middleware.Rate(rest.NewISearchHandler(log, searchClient), cfg.SearchRate)))
	mux.Handle("POST /api/login", middleware.Metrics(rest.NewLoginHandler(log, verifier)))
	mux.Handle("GET /metrics", rest.NewMetricsHandler())

	server := http.Server{
		Addr:        cfg.HTTPConfig.Address,
		ReadTimeout: cfg.HTTPConfig.Timeout,
		Handler:     mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	chanError := make(chan error, 1)
	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		ctxShutdown, cancel := context.WithTimeout(context.Background(), gracefulShutdownTime)
		defer cancel()

		chanError <- server.Shutdown(ctxShutdown)
	}()

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Error("server closed unexpectedly", "error", err)
		return
	}

	// Make sure the program doesn't exit and waits instead for Shutdown to return.
	if err = <-chanError; err != nil {
		log.Debug("hard stop server", "err", err)
	}
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
