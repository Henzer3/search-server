package db

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"
	"yadro.com/course/update/core"
)

func TestStorage_Add(t *testing.T) {
	conn, mock, err := sqlxmock.Newx()
	require.NoError(t, err)
	defer func() {
		err = conn.Close()
		require.NoError(t, err)
	}()

	db := DB{
		log:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		conn: conn,
	}

	type args struct {
		ID  int
		Url string
	}

	a := args{ID: 1, Url: "https://example.com/1.jpg"}

	type mockBehavior func(args)

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		wantErr      bool
	}{
		{
			name: "Add_OK",
			mockBehavior: func(args args) {
				mock.ExpectBegin()

				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO comics (num, img_url)")).WithArgs(args.ID, args.Url).WillReturnResult(sqlxmock.NewResult(1, 1))
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO words (word, comics_num)")).WithArgs(sqlxmock.AnyArg(), sqlxmock.AnyArg()).WillReturnResult(sqlxmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name: "Add_Error",
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO comics (num, img_url)").WithArgs(args.ID, args.Url).WillReturnError(errors.New("insert error"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Add_Error_Second",
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO comics (num, img_url)").WithArgs(args.ID, args.Url).WillReturnResult(sqlxmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO words (word, comics_num)").WithArgs(sqlxmock.AnyArg(), sqlxmock.AnyArg()).WillReturnError(errors.New("insert error"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(a)
			err := db.Add(context.Background(), core.Comics{ID: 1, URL: "https://example.com/1.jpg", Words: []string{"golang"}})
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}

}

func TestStorage_Stats(t *testing.T) {
	conn, mock, err := sqlxmock.Newx()
	require.NoError(t, err)
	defer func() {
		err = conn.Close()
		require.NoError(t, err)
	}()

	db := DB{
		log:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		conn: conn,
	}

	type mockBehavior func()

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		wantErr      bool
	}{
		{
			name: "Stats_OK",
			mockBehavior: func() {
				row := sqlxmock.NewRows([]string{"words_total", "words_unique", "comics_fetched"}).AddRow(2, 2, 2)
				mock.ExpectQuery("SELECT").WillReturnRows(row)
			},
		},
		{
			name: "Stats_Error",
			mockBehavior: func() {
				mock.ExpectQuery("SELECT").WillReturnError(errors.New("Query error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()
			res, err := db.Stats(context.Background())
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, core.DBStats{ComicsFetched: 2, WordsTotal: 2, WordsUnique: 2}, res)
		})
	}

}

func TestStorage_Ids(t *testing.T) {
	conn, mock, err := sqlxmock.Newx()
	require.NoError(t, err)
	defer func() {
		err = conn.Close()
		require.NoError(t, err)
	}()

	db := DB{
		log:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		conn: conn,
	}

	type mockBehavior func()

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		wantErr      bool
	}{
		{
			name: "Ids_OK",
			mockBehavior: func() {
				row := sqlxmock.NewRows([]string{"num"}).AddRow(1).AddRow(2)
				mock.ExpectQuery("SELECT num FROM comics ORDER BY num").WillReturnRows(row)
			},
		},
		{
			name: "Ids_Error",
			mockBehavior: func() {
				mock.ExpectQuery("SELECT num FROM comics ORDER BY num").WillReturnError(errors.New("Query error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()
			res, err := db.IDs(context.Background())
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, []int{1, 2}, res)
		})
	}

}

func TestStorage_Drop(t *testing.T) {
	conn, mock, err := sqlxmock.Newx()
	require.NoError(t, err)
	defer func() {
		err = conn.Close()
		require.NoError(t, err)
	}()

	db := DB{
		log:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		conn: conn,
	}

	type mockBehavior func()

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		wantErr      bool
	}{
		{
			name: "Drop_OK",
			mockBehavior: func() {
				mock.ExpectExec("TRUNCATE TABLE words, comics RESTART IDENTITY CASCADE").WillReturnResult(sqlxmock.NewResult(1, 1))
			},
		},
		{
			name: "Ids_Error",
			mockBehavior: func() {
				mock.ExpectExec("TRUNCATE TABLE words, comics RESTART IDENTITY CASCADE").WillReturnError(errors.New("Drop error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()
			err := db.Drop(context.Background())
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}

}
