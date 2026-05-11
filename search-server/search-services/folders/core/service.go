package core

import (
	"context"
	"errors"
	"log/slog"

	"yadro.com/course/folders/entity"
)

type service struct {
	comicsDB entity.ComicsDB
	folderDB entity.FolderDB
	search   entity.Search
	log      *slog.Logger
}

func New(log *slog.Logger, cdb entity.ComicsDB, fdb entity.FolderDB, search entity.Search) *service {
	return &service{
		log:      log,
		search:   search,
		comicsDB: cdb,
		folderDB: fdb,
	}
}

func (s *service) CreateFolder(ctx context.Context, userID int64, name string) (int64, error) {
	id, err := s.folderDB.CreateFolder(ctx, userID, name)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *service) DeleteFolder(ctx context.Context, userID int64, folderID int64) error {
	if err := s.checkPermission(ctx, userID, folderID); err != nil {
		return err
	}

	return s.folderDB.DeleteFolder(ctx, folderID)
}

func (s *service) AddComics(ctx context.Context, in entity.AddComicsIn) error {
	if err := s.checkPermission(ctx, in.UserID, in.FolderId); err != nil {
		return err
	}

	comics, err := s.search.GetComics(ctx, in.ComicsID)
	if err != nil {
		if errors.Is(err, entity.ErrComicsNotExist) {
			return entity.ErrComicsNotExist
		}
		s.log.Error("cant GetComics in core", "err", err)
		return err
	}

	addInfo := entity.AddComicInfo{
		FolderID: in.FolderId,
		ComicsID: comics.ComicsID,
		URL:      comics.URL,
	}

	if err := s.comicsDB.AddComics(ctx, addInfo); err != nil {
		if errors.Is(err, entity.ErrComicsExist) {
			return err
		}
		s.log.Error("cant AddComics in core", "err", err)
		return err
	}

	return nil
}

func (s *service) DeleteComics(ctx context.Context, in entity.DeleteComicsIn) error {
	if err := s.checkPermission(ctx, in.UserID, in.FolderId); err != nil {
		return err
	}
	if err := s.comicsDB.DeleteComics(ctx, in.FolderId, in.ComicsID); err != nil {
		if errors.Is(err, entity.ErrComicsNotExist) {
			return entity.ErrComicsNotExist
		}
		s.log.Error("cant DeleteComics in core", "err", err)
		return err
	}

	return nil
}

func (s *service) ListComics(ctx context.Context, userID int64, folderID int64) ([]entity.Comics, error) {
	if err := s.checkPermission(ctx, userID, folderID); err != nil {
		return nil, err
	}

	comics, err := s.comicsDB.ListComics(ctx, folderID)
	if err != nil {
		s.log.Error("cant get list of comics in core", "err", err)
		return nil, err
	}

	return comics, nil
}

func (s *service) ListFolders(ctx context.Context, userID int64) ([]entity.Folder, error) {
	folders, err := s.folderDB.ListFolders(ctx, userID)
	if err != nil {
		s.log.Error("cant get list of folders in core", "err", err)
		return nil, err
	}
	return folders, nil
}

func (s *service) checkPermission(ctx context.Context, userID int64, folderID int64) error {
	id, err := s.folderDB.FolderOwnerID(ctx, folderID)
	if err != nil {
		if errors.Is(err, entity.ErrFolderNotExist) {
			return err
		}
		s.log.Error("cant getUser in core", "err", err)
		return err
	}

	if id != userID {
		return entity.ErrNoPermission
	}

	return nil
}
