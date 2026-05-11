package core

import "context"

//go:generate go run go.uber.org/mock/mockgen -source=ports.go -destination=mocks/mock.go -package=mocks
type Normalizer interface {
	Norm(context.Context, string) ([]string, error)
}

type Pinger interface {
	Ping(context.Context) error
}

type Updater interface {
	Update(context.Context) error
	Stats(context.Context) (UpdateStats, error)
	Status(context.Context) (UpdateStatus, error)
	Drop(context.Context) error
}

type Searcher interface {
	Search(context.Context, string, int) ([]ImageInformation, error)
	ISearch(context.Context, string, int) ([]ImageInformation, error)
}

type Authenticator interface {
	Login(ctx context.Context, in LoginRequest) ([]byte, error)
	Register(ctx context.Context, email string, password string) (int64, error)
	Verify(ctx context.Context, token string) (UserPermissions, error)
}

type Folderer interface {
	CreateFolder(ctx context.Context, uid int64, name string) (int64, error)
	DeleteFolder(ctx context.Context, uid int64, fid int64) error
	AddComics(ctx context.Context, uid int64, fid int64, cid int64) error
	DeleteComics(ctx context.Context, uid int64, fid int64, cid int64) error
	ListComics(ctx context.Context, uid int64, fid int64) ([]ImageInformation, error)
	ListFolders(ctx context.Context, uid int64) ([]Folder, error)
}
