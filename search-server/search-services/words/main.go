package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	wordspb "yadro.com/course/proto/words"

	"yadro.com/course/words/config"
	"yadro.com/course/words/handler"
)

const gracefulShutdownTime = 2
const defaultConfigFileName = "config.yaml"

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", defaultConfigFileName, "server configuration file")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("cant load config", "err", err)
		return
	}

	port := ":" + cfg.Port

	listener, err := net.Listen("tcp", port)
	if err != nil {
		logger.Error("failed to listen", "err", err)
		return
	}

	server := grpc.NewServer()
	wordspb.RegisterWordsServer(server, handler.NewServer(logger))
	reflection.Register(server)

	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Serve(listener)
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-signalCtx.Done():
	case err := <-errChan:
		logger.Error("failed to serve", "err", err)
		return
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
		logger.Error("failed to serve", "err", err)
		return
	}

}
