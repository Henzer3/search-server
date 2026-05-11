package db

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"
)

func TestStorage_Search(t *testing.T) {
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
			name: "Search_OK",
			mockBehavior: func() {
				rows := sqlxmock.NewRows([]string{"num", "img_url"}).
					AddRow(1, "https://example.com/1.jpg").AddRow(2, "https://example.com/2.jpg")

				mock.ExpectQuery("SELECT comics.num, comics.img_url").WithArgs(sqlxmock.AnyArg()).WillReturnRows(rows)
			},
		},
		{
			name: "Search_Error",
			mockBehavior: func() {
				mock.ExpectQuery("SELECT comics.num, comics.img_url").WithArgs(sqlxmock.AnyArg()).WillReturnError(errors.New("some error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()
			res, err := db.Search(context.Background(), []string{"golang", "goffer"})
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, 2, len(res))
		})
	}

}

func TestStorage_CreateIndex(t *testing.T) {
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
			name: "Create_Error",
			mockBehavior: func() {
				mock.ExpectQuery("SELECT words.word, comics.num, comics.img_url").WillReturnError(errors.New("some error"))
			},
			wantErr: true,
		},
		{
			name: "Create_OK",
			mockBehavior: func() {
				rows := sqlxmock.NewRows([]string{"word", "num", "img_url"}).
					AddRow("apple", 2, "https://example.com/2.jpg").
					AddRow("golang", 1, "https://example.com/1.jpg").
					AddRow("golang", 3, "https://example.com/3.jpg")

				mock.ExpectQuery("SELECT words.word, comics.num, comics.img_url").WillReturnRows(rows)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()
			res, err := db.CreateIndex()
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, 3, len(res))
			require.Equal(t, "apple", res[0].Word)
			require.Equal(t, 1, res[1].ID)
		})
	}

}
