package grpc

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/folders/entity"
	folderspb "yadro.com/course/proto/folders"
)

type foldersServer struct {
	folderspb.UnimplementedFoldersServer
	log     *slog.Logger
	service entity.Folderer
}

func NewServer(log *slog.Logger, service entity.Folderer) *foldersServer {
	return &foldersServer{
		log:     log,
		service: service,
	}
}

func (s *foldersServer) CreateFolder(ctx context.Context, in *folderspb.CreateFolderRequest) (*folderspb.CreateFolderResponse, error) {
	folderID, err := s.service.CreateFolder(ctx, in.GetUserId(), in.GetName())
	if err != nil {
		if errors.Is(err, entity.ErrFolderExist) {
			return nil, status.Error(codes.AlreadyExists, "folder already exists")
		}
		s.log.Error("cant CreateFolder in grpc", "err", err)
		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &folderspb.CreateFolderResponse{FolderId: folderID}, nil
}

func (s *foldersServer) DeleteFolder(ctx context.Context, in *folderspb.DeleteFolderRequest) (*emptypb.Empty, error) {
	if err := s.service.DeleteFolder(ctx, in.GetUserId(), in.GetFolderId()); err != nil {
		if errors.Is(err, entity.ErrFolderNotExist) {
			return nil, status.Error(codes.FailedPrecondition, "folder doesnt exist")
		} else if errors.Is(err, entity.ErrNoPermission) {
			return nil, status.Error(codes.PermissionDenied, "no permissions")
		}
		s.log.Error("cant DeleteFolder in grpc", "err", err)
		return nil, status.Error(codes.Internal, "Internal error")
	}
	return nil, nil
}

func (s *foldersServer) AddComics(ctx context.Context, in *folderspb.AddComicsRequest) (*emptypb.Empty, error) {
	if err := s.service.AddComics(ctx, entity.AddComicsIn{UserID: in.GetUserId(), FolderId: in.GetFolderId(), ComicsID: in.GetComicsId()}); err != nil {
		switch {
		case errors.Is(err, entity.ErrComicsExist):
			return nil, status.Error(codes.AlreadyExists, "already exist")
		case errors.Is(err, entity.ErrComicsNotExist):
			return nil, status.Error(codes.FailedPrecondition, "comics doesnt exist")
		case errors.Is(err, entity.ErrNoPermission):
			return nil, status.Error(codes.PermissionDenied, "no permissions")
		case errors.Is(err, entity.ErrFolderNotExist):
			return nil, status.Error(codes.FailedPrecondition, "folder doesnt exist")
		}
		s.log.Error("cant AddComics in grpc", "err", err)
		return nil, status.Error(codes.Internal, "Internal error")
	}
	return nil, nil
}

func (s *foldersServer) DeleteComics(ctx context.Context, in *folderspb.DeleteComicsRequest) (*emptypb.Empty, error) {
	if err := s.service.DeleteComics(ctx, entity.DeleteComicsIn{UserID: in.GetUserId(), FolderId: in.GetFolderId(), ComicsID: in.GetComicsId()}); err != nil {
		switch {
		case errors.Is(err, entity.ErrComicsNotExist):
			return nil, status.Error(codes.FailedPrecondition, "comics doesnt exist")
		case errors.Is(err, entity.ErrNoPermission):
			return nil, status.Error(codes.PermissionDenied, "no permissions")
		case errors.Is(err, entity.ErrFolderNotExist):
			return nil, status.Error(codes.FailedPrecondition, "folder doesnt exist")
		}
		s.log.Error("cant DeleteComics in grpc", "err", err)
		return nil, status.Error(codes.Internal, "Internal error")
	}
	return nil, nil
}

func (s *foldersServer) ListComics(ctx context.Context, in *folderspb.ListComicsRequest) (*folderspb.ListComicsResponse, error) {
	comics, err := s.service.ListComics(ctx, in.GetUserId(), in.GetFolderId())
	if err != nil {
		if errors.Is(err, entity.ErrFolderNotExist) {
			return nil, status.Error(codes.FailedPrecondition, "folder doesnt exist")
		} else if errors.Is(err, entity.ErrNoPermission) {
			return nil, status.Error(codes.PermissionDenied, "no permissions")
		}
		s.log.Error("cant ListComics in grpc", "err", err)
		return nil, status.Error(codes.Internal, "Internal error")
	}

	res := make([]*folderspb.Comics, 0, len(comics))

	for _, v := range comics {
		res = append(res, &folderspb.Comics{ComicsId: v.ComicsID, Url: v.URL})
	}

	return &folderspb.ListComicsResponse{Comics: res}, nil
}

func (s *foldersServer) ListFolders(ctx context.Context, in *folderspb.ListFoldersRequest) (*folderspb.ListFoldersResponse, error) {
	folders, err := s.service.ListFolders(ctx, in.GetUserId())
	if err != nil {
		s.log.Error("cant ListFolders in grpc", "err", err)
		return nil, status.Error(codes.Internal, "Internal error")
	}

	res := make([]*folderspb.Folder, 0, len(folders))

	for _, v := range folders {
		res = append(res, &folderspb.Folder{FolderId: v.FolderID, Name: v.Name})
	}

	return &folderspb.ListFoldersResponse{Folders: res}, nil
}
