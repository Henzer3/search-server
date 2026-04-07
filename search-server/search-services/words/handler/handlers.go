package handler

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	wordspb "yadro.com/course/proto/words"
	"yadro.com/course/words/words"
)

const maxMessageSize = 4 * 1024

type server struct {
	wordspb.UnimplementedWordsServer
	logger *slog.Logger
}

func NewServer(log *slog.Logger) *server {
	return &server{logger: log}
}

func (s *server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return new(emptypb.Empty), nil
}

func (s *server) Norm(ctx context.Context, in *wordspb.WordsRequest) (*wordspb.WordsReply, error) {
	if len(in.Phrase) > maxMessageSize {
		err := status.Error(codes.ResourceExhausted, "too much message")
		s.logger.Error("too much messages", "err", err)
		return nil, err
	}

	ans, err := words.StemSlice(in.Phrase, true)
	if err != nil {
		s.logger.Error("too much messages", "err", err)
		return nil, status.Error(codes.Internal, "processing failed")
	}
	return &wordspb.WordsReply{Words: ans}, nil
}
