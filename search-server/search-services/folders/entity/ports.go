package entity

import (
	"context"
)

type Folderer interface {
	CreateFolder(ctx context.Context, userID int64, name string) (int64, error)
	DeleteFolder(ctx context.Context, userID int64, folderID int64) error
	AddComics(ctx context.Context, in AddComicsIn) error
	DeleteComics(ctx context.Context, in DeleteComicsIn) error
	ListComics(ctx context.Context, userID int64, folderID int64) ([]Comics, error)
	ListFolders(ctx context.Context, userID int64) ([]Folder, error)
}

type Search interface {
	GetComics(ctx context.Context, id int64) (Comics, error)
}

type ComicsDB interface {
	AddComics(ctx context.Context, in AddComicInfo) error
	DeleteComics(ctx context.Context, folderID int64, comicsID int64) error
	ListComics(ctx context.Context, folderID int64) ([]Comics, error)
}

type FolderDB interface {
	CreateFolder(ctx context.Context, userID int64, name string) (int64, error)
	DeleteFolder(ctx context.Context, folderID int64) error
	ListFolders(ctx context.Context, userID int64) ([]Folder, error)
	FolderOwnerID(ctx context.Context, folderID int64) (int64, error)
}
