package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	// "github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"yadro.com/course/folders/entity"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func New(log *slog.Logger, address string) (*DB, error) {
	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	return &DB{
		log:  log,
		conn: db,
	}, nil
}

func (db *DB) Close() error {
	if err := db.conn.Close(); err != nil {
		db.log.Error("cant close db conn", "err", err)
		return err
	}
	return nil
}

func (db *DB) CreateFolder(ctx context.Context, userID int64, name string) (int64, error) {
	var id int64
	if err := db.conn.GetContext(ctx, &id, `
	INSERT INTO folder.folders (name, user_id)
	VALUES ($1, $2)
	RETURNING id
	`, name, userID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, entity.ErrFolderExist
		}

		db.log.Error("cant CreateFolder in db", "err", err)
		return 0, err
	}

	return id, nil
}

func (db *DB) DeleteFolder(ctx context.Context, folderID int64) error {
	res, err := db.conn.ExecContext(ctx, `
	DELETE FROM folder.folders
	WHERE id = $1
	`, folderID)

	if err != nil {
		db.log.Error("cant deleteFolder in db", "err", err)
		return err
	}

	deleteRows, err := res.RowsAffected()

	if err != nil {
		db.log.Error("cant get affected rows in DeleteFolder", "err", err)
		return err
	}

	if deleteRows == 0 {
		return entity.ErrFolderNotExist
	}

	return nil
}

func (db *DB) FolderOwnerID(ctx context.Context, folderID int64) (int64, error) {
	var id int64

	if err := db.conn.GetContext(ctx, &id, `
	SELECT user_id FROM folder.folders
	WHERE id = $1
	`, folderID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, entity.ErrFolderNotExist
		}
		db.log.Error("cant GetUser in db", "err", err)
		return 0, err
	}

	return id, nil
}

type folder struct {
	FolderID int64  `db:"id"`
	Name     string `db:"name"`
}

func (db *DB) ListFolders(ctx context.Context, userID int64) ([]entity.Folder, error) {
	var folders []folder

	if err := db.conn.SelectContext(ctx, &folders, `
	SELECT id, name FROM folder.folders
	WHERE user_id = $1
	`, userID); err != nil {
		db.log.Error("cant get ListFolders in db", "err", err)
		return nil, err
	}

	res := make([]entity.Folder, 0, len(folders))
	for _, v := range folders {
		res = append(res, entity.Folder{FolderID: v.FolderID, Name: v.Name})
	}

	return res, nil
}

func (db *DB) AddComics(ctx context.Context, in entity.AddComicInfo) error {
	_, err := db.conn.ExecContext(ctx, `
	INSERT INTO folder.folder_comics (comic_id, url, folder_id)
	VALUES ($1, $2, $3)
	`, in.ComicsID, in.URL, in.FolderID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entity.ErrComicsExist
		}

		db.log.Error("cant AddComics in db", "err", err)
		return err
	}

	return nil
}

func (db *DB) DeleteComics(ctx context.Context, folderID int64, comicsID int64) error {
	res, err := db.conn.ExecContext(ctx, `
	DELETE FROM folder.folder_comics
	WHERE folder_id = $1 AND comic_id = $2
	`, folderID, comicsID)

	if err != nil {
		db.log.Error("cant DeleteComic in db", "err", err)
		return err
	}

	deletedRows, err := res.RowsAffected()

	if err != nil {
		db.log.Error("cant get affected rows in DeleteFolder", "err", err)
		return err
	}

	if deletedRows == 0 {
		return entity.ErrComicsNotExist
	}

	return nil
}

type comic struct {
	ComicID int64  `db:"comic_id"`
	URL     string `db:"url"`
}

func (db *DB) ListComics(ctx context.Context, folderID int64) ([]entity.Comics, error) {
	var comics []comic

	if err := db.conn.SelectContext(ctx, &comics, `
	SELECT comic_id, url FROM folder.folder_comics
	WHERE folder_id = $1
	`, folderID); err != nil {
		db.log.Error("cant get ListComics", "err", err)
		return nil, err
	}

	res := make([]entity.Comics, 0, len(comics))
	for _, v := range comics {
		res = append(res, entity.Comics{ComicsID: v.ComicID, URL: v.URL})
	}

	return res, nil
}
