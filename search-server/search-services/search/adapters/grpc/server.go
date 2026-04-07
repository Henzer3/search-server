package grpc

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/core"
)

type Server struct {
	searchpb.UnimplementedSearchServer
	log     *slog.Logger
	service core.Searcher
}

func NewServer(log *slog.Logger, service core.Searcher) *Server {
	return &Server{log: log, service: service}
}

func (s *Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return new(emptypb.Empty), nil
}

func (s *Server) Search(ctx context.Context, in *searchpb.SearchRequest) (*searchpb.SearchReply, error) {
	out, err := s.search(ctx, in.Phrase, int(in.Limit), s.service.Search)
	if err != nil {
		s.log.Error("cant search in db", "err", err, "phrase", in.Phrase)
		return nil, err
	}
	return out, nil
}

func (s *Server) ISearch(ctx context.Context, in *searchpb.SearchRequest) (*searchpb.SearchReply, error) {
	out, err := s.search(ctx, in.Phrase, int(in.Limit), s.service.ISearch)
	if err != nil {
		s.log.Error("cant Isearch in index", "err", err, "phrase", in.Phrase)
		return nil, err
	}
	return out, nil
}

func (s *Server) search(ctx context.Context, phrase string, limit int, search func(ctx context.Context, phrase string, limit int) ([]core.ImageInformation, error)) (*searchpb.SearchReply, error) {
	out, err := search(ctx, phrase, limit)
	if err != nil {
		s.log.Error("Cant search phrase", "err", err, "string", phrase)
		return nil, status.Error(codes.Internal, "Internal error")
	}
	var images []*searchpb.Image
	for _, v := range out {
		images = append(images, &searchpb.Image{Id: int64(v.ID), Url: v.Url})
	}
	return &searchpb.SearchReply{Images: images}, nil
}
