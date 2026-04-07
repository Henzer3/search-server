package grpc

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	updatepb "yadro.com/course/proto/update"
	"yadro.com/course/update/core"
)

type Server struct {
	updatepb.UnimplementedUpdateServer
	service core.Updater
	log     *slog.Logger
}

func NewServer(log *slog.Logger, service core.Updater) *Server {
	return &Server{log: log, service: service}
}

func (s *Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return new(emptypb.Empty), nil
}

func (s *Server) Status(ctx context.Context, _ *emptypb.Empty) (*updatepb.StatusReply, error) {
	var status updatepb.Status
	switch s.service.Status(ctx) {
	case core.StatusRunning:
		status = updatepb.Status_STATUS_RUNNING
	case core.StatusIdle:
		status = updatepb.Status_STATUS_IDLE
	default:
		status = updatepb.Status_STATUS_UNSPECIFIED
	}
	return &updatepb.StatusReply{Status: status}, nil
}

func (s *Server) Update(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.service.Update(ctx); err != nil {
		if errors.Is(err, core.ErrAlreadyUpdating) {
			return nil, status.Error(codes.Unavailable, "update already in progress")
		}
		s.log.Error("cant update", "err", err)
		return nil, status.Error(codes.Internal, "processing failed")
	}
	return new(emptypb.Empty), nil
}

func (s *Server) Stats(ctx context.Context, _ *emptypb.Empty) (*updatepb.StatsReply, error) {
	stats, err := s.service.Stats(ctx)
	if err != nil {
		s.log.Error("cant get stats", "err", err)
		return nil, status.Error(codes.Internal, "processing failed")
	}
	return &updatepb.StatsReply{WordsTotal: int64(stats.WordsTotal), WordsUnique: int64(stats.WordsUnique),
		ComicsFetched: int64(stats.ComicsFetched), ComicsTotal: int64(stats.ComicsTotal)}, nil
}

func (s *Server) Drop(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.service.Drop(ctx); err != nil {
		return nil, status.Error(codes.Internal, "process droping failed")
	}
	return new(emptypb.Empty), nil
}
