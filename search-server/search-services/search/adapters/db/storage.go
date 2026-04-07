package db

import (
	"context"
	"log/slog"

	"yadro.com/course/search/core"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

type imageInformation struct {
	ID  int    `db:"num"`
	Url string `db:"img_url"`
}

type wordInformation struct {
	Word string `db:"word"`
	ID   int    `db:"num"`
	Url  string `db:"img_url"`
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
	return db.conn.Close()
}

func (db *DB) Search(ctx context.Context, words []string) ([]core.ImageInformation, error) {
	if len(words) == 0 {
		return nil, nil
	}

	var comics []imageInformation

	const query = `
		SELECT comics.num, comics.img_url
		FROM words
		JOIN comics ON words.comics_num = comics.num
		WHERE words.word = ANY($1::text[])
		GROUP BY comics.num, comics.img_url
		ORDER BY COUNT(DISTINCT words.word) DESC, comics.num`

	if err := db.conn.SelectContext(ctx, &comics, query, pq.Array(words)); err != nil {
		db.log.Error("cant select", "err", err)
		return nil, err
	}

	coreComics := make([]core.ImageInformation, 0, len(comics))
	for _, v := range comics {
		coreComics = append(coreComics, core.ImageInformation{ID: v.ID, Url: v.Url})
	}
	return coreComics, nil
}

func (db *DB) CreateIndex() ([]core.WordInformation, error) {
	const query = `
		SELECT words.word, comics.num, comics.img_url
		FROM words
		JOIN comics ON comics.num = words.comics_num
		ORDER BY words.word, comics.num`

	var words []wordInformation

	if err := db.conn.Select(&words, query); err != nil {
		db.log.Error("cant select for create index", "err", err)
		return nil, err
	}

	wordsInfo := make([]core.WordInformation, 0, len(words))
	for _, v := range words {
		wordsInfo = append(wordsInfo, core.WordInformation{Word: v.Word, ID: v.ID, Url: v.Url})
	}
	return wordsInfo, nil
}
