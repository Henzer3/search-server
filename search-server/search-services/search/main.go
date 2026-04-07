package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/adapters/db"
	searchgrpc "yadro.com/course/search/adapters/grpc"
	"yadro.com/course/search/adapters/initiator"
	"yadro.com/course/search/adapters/inmemory"
	sub "yadro.com/course/search/adapters/nats"
	"yadro.com/course/search/adapters/words"
	"yadro.com/course/search/config"
	"yadro.com/course/search/core"

	"github.com/nats-io/nats.go"
)

const gracefulShutdownTime = 2

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()
	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.LogLevel)

	if err := run(cfg, log); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	log.Info("starting server")
	log.Debug("debug messages are enabled")

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	// database adapter
	storage, err := db.New(log, cfg.DBAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %v", err)
	}

	defer func() {
		if err := storage.Close(); err != nil {
			logger.Error("close conn in database adapter", "err", err)
		}
	}()

	// words adapter
	words, err := words.NewClient(cfg.WordsAddress, log)
	if err != nil {
		return fmt.Errorf("failed create Words client: %v", err)
	}
	defer func() {
		if err := words.Close(); err != nil {
			logger.Error("close conn in words adapter", "err", err)
		}
	}()

	// create inMemory Repository
	inMemoryRepository := inmemory.NewRep(log)

	// service
	searcher := core.NewService(log, storage, words, inMemoryRepository)

	// initiator
	initiator, err := initiator.NewInitiator(log, cfg.IndexTtl, func() error {
		return searcher.RebuildIndex()
	})

	if err != nil {
		log.Error("cant create initiator", "err", err)
		return fmt.Errorf("failed create initiator: %v", err)
	}

	defer initiator.Stop()

	// subscriber
	subscriber, err := sub.New(log, cfg.NatsAdress, 10)
	if err != nil {
		return fmt.Errorf("failed create subscriber: %v", err)
	}
	defer func() {
		if err := subscriber.Stop(); err != nil {
			log.Error("cant stop subscriber", "err", err)
		}
	}()

	err = subscriber.Subscribe("xkcd.db.updated", func(_ *nats.Msg) {
		if err := searcher.RebuildIndex(); err != nil {
			log.Error("cant rebuild index", "err", err)
		}
	})

	if err != nil {
		return fmt.Errorf("cant subscribe on xkcd.db.updated: %v", err)
	}

	err = subscriber.Subscribe("xkcd.db.deleted", func(_ *nats.Msg) {
		searcher.DeleteIndex()
	})

	if err != nil {
		return fmt.Errorf("cant subscribe on xkcd.db.deleted: %v", err)
	}

	subscriber.StartListen()

	// grpc server
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	searchpb.RegisterSearchServer(server, searchgrpc.NewServer(logger, searcher))
	reflection.Register(server)

	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Serve(listener)
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	select {
	case <-signalCtx.Done():
	case err := <-errChan:
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*gracefulShutdownTime)
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-ctx.Done():
		logger.Error("time is up")
		server.Stop()
	}

	if err := <-errChan; !errors.Is(err, grpc.ErrServerStopped) {
		return err
	}
	return nil
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
