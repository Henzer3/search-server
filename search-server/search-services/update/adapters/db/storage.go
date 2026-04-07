package db

import (
	"context"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"yadro.com/course/update/core"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

type DBStats struct {
	WordsTotal    int `db:"words_total"`
	WordsUnique   int `db:"words_unique"`
	ComicsFetched int `db:"comics_fetched"`
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

func (db *DB) Add(ctx context.Context, comics core.Comics) error {
	tx, err := db.conn.BeginTxx(ctx, nil)
	if err != nil {
		db.log.Error("BgeinTxx error", "err", err)
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	_, err = tx.ExecContext(ctx, `
	INSERT INTO comics (num, img_url)
	VALUES ($1, $2)
	ON CONFLICT (num) DO NOTHING
	`, comics.ID, comics.URL)
	if err != nil {
		db.log.Error("cant add in comics", "err", err, "num", comics.ID)
		return err
	}

	if len(comics.Words) > 0 {
		comicsNums := make([]int, len(comics.Words))
		for i := range comics.Words {
			comicsNums[i] = comics.ID
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO words (word, comics_num)
			SELECT *
			FROM unnest($1::text[], $2::integer[])
			ON CONFLICT (word, comics_num) DO NOTHING
		`, pq.Array(comics.Words), pq.Array(comicsNums))
		if err != nil {
			db.log.Error("cant add in words", "err", err, "num", comics.ID)
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) Stats(ctx context.Context) (core.DBStats, error) {
	var DBStats DBStats
	if err := db.conn.GetContext(ctx, &DBStats, `
	SELECT 
		COUNT(*)  as words_total,
		COUNT(DISTINCT word) as words_unique,
		(SELECT COUNT(*) FROM comics) as comics_fetched
	FROM words`); err != nil {
		db.log.Error("cant get count of words", "err", err)
		return core.DBStats{}, err
	}

	return core.DBStats{WordsTotal: DBStats.WordsTotal, WordsUnique: DBStats.WordsUnique, ComicsFetched: DBStats.ComicsFetched}, nil
}

func (db *DB) IDs(ctx context.Context) ([]int, error) {
	ids := make([]int, 0)
	if err := db.conn.SelectContext(ctx, &ids, `SELECT num FROM comics ORDER BY num`); err != nil {
		db.log.Error("cant get ids", "err", err)
		return nil, err
	}
	return ids, nil
}

func (db *DB) Drop(ctx context.Context) error {
	if _, err := db.conn.ExecContext(ctx, `
	TRUNCATE TABLE words, comics RESTART IDENTITY CASCADE`); err != nil {
		db.log.Error("cant delete tables", "err", err)
		return err
	}

	return nil
}
