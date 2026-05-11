package core_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"yadro.com/course/search/core"
	"yadro.com/course/search/core/mocks"
)

var logger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestService(t *testing.T) {
	t.Parallel()
	const (
		phrase = "I love golang so much"
		limit  = 20
	)

	tests := []struct {
		name      string
		mockSetup func(db *mocks.MockDB, in *mocks.MockInMemoryRep, words *mocks.MockWords)
		call      func(s *core.Service) ([]core.ImageInformation, error)
		wantErr   bool
	}{
		{
			name: "Search_OK",
			mockSetup: func(db *mocks.MockDB, in *mocks.MockInMemoryRep, words *mocks.MockWords) {
				words.EXPECT().Norm(gomock.Any(), phrase).Return([]string{"I", "love", "golang"}, nil).Times(1)

				db.EXPECT().Search(gomock.Any(), []string{"I", "love", "golang"}).Return([]core.ImageInformation{
					{ID: 1, Url: "https://example.com/1.jpg"},
					{ID: 2, Url: "https://example.com/2.jpg"},
					{ID: 3, Url: "https://example.com/3.jpg"},
				}, nil).Times(1)

			},
			call: func(s *core.Service) ([]core.ImageInformation, error) {
				return s.Search(context.Background(), phrase, limit)
			},
		},

		{
			name: "Search_NormError",
			mockSetup: func(db *mocks.MockDB, in *mocks.MockInMemoryRep, words *mocks.MockWords) {
				words.EXPECT().Norm(gomock.Any(), phrase).Return(nil, errors.New("Norm error")).Times(1)
			},
			call: func(s *core.Service) ([]core.ImageInformation, error) {
				return s.Search(context.Background(), phrase, limit)
			},
			wantErr: true,
		},

		{
			name: "Search_DB_Error",
			mockSetup: func(db *mocks.MockDB, in *mocks.MockInMemoryRep, words *mocks.MockWords) {
				words.EXPECT().Norm(gomock.Any(), phrase).Return([]string{"I", "love", "golang"}, nil).Times(1)

				db.EXPECT().Search(gomock.Any(), []string{"I", "love", "golang"}).Return(nil, errors.New("DB error")).Times(1)

			},
			call: func(s *core.Service) ([]core.ImageInformation, error) {
				return s.Search(context.Background(), phrase, limit)
			},
			wantErr: true,
		},

		{
			name: "ISearch_OK",
			mockSetup: func(db *mocks.MockDB, in *mocks.MockInMemoryRep, words *mocks.MockWords) {
				words.EXPECT().Norm(gomock.Any(), phrase).Return([]string{"I", "love", "golang"}, nil).Times(1)

				in.EXPECT().Search(gomock.Any(), []string{"I", "love", "golang"}).Return([]core.QuantityComics{
					{
						ImageInfo: core.ImageInformation{
							ID:  3,
							Url: "https://example.com/3.jpg",
						},
						Total: 1,
					},
					{
						ImageInfo: core.ImageInformation{
							ID:  2,
							Url: "https://example.com/2.jpg",
						},
						Total: 5,
					},
					{
						ImageInfo: core.ImageInformation{
							ID:  1,
							Url: "https://example.com/1.jpg",
						},
						Total: 5,
					},
				}, nil).Times(1)

			},
			call: func(s *core.Service) ([]core.ImageInformation, error) {
				return s.ISearch(context.Background(), phrase, limit)
			},
		},

		{
			name: "ISearch_NormError",
			mockSetup: func(db *mocks.MockDB, in *mocks.MockInMemoryRep, words *mocks.MockWords) {
				words.EXPECT().Norm(gomock.Any(), phrase).Return(nil, errors.New("Norm error")).Times(1)
			},
			call: func(s *core.Service) ([]core.ImageInformation, error) {
				return s.ISearch(context.Background(), phrase, limit)
			},
			wantErr: true,
		},

		{
			name: "Search_InMemory_Error",
			mockSetup: func(db *mocks.MockDB, in *mocks.MockInMemoryRep, words *mocks.MockWords) {
				words.EXPECT().Norm(gomock.Any(), phrase).Return([]string{"I", "love", "golang"}, nil).Times(1)

				in.EXPECT().Search(gomock.Any(), []string{"I", "love", "golang"}).Return(nil, errors.New("InMemory error")).Times(1)

			},
			call: func(s *core.Service) ([]core.ImageInformation, error) {
				return s.ISearch(context.Background(), phrase, limit)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockDB(ctrl)
			inmemory := mocks.NewMockInMemoryRep(ctrl)
			words := mocks.NewMockWords(ctrl)

			tc.mockSetup(db, inmemory, words)

			service := core.NewService(logger, db, words, inmemory)

			info, err := tc.call(service)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, info)
				return
			}
			require.NoError(t, err)
			require.Equal(t, []core.ImageInformation{
				{ID: 1, Url: "https://example.com/1.jpg"},
				{ID: 2, Url: "https://example.com/2.jpg"},
				{ID: 3, Url: "https://example.com/3.jpg"},
			}, info)

		})
	}

}

func Test_Service_RebuildIndex(t *testing.T) {
	t.Parallel()

	type mockBehavior func(
		db *mocks.MockDB,
		in *mocks.MockInMemoryRep,
		words *mocks.MockWords,
	)

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		wantErr      bool
	}{
		{
			name: "Succses",
			mockBehavior: func(
				db *mocks.MockDB,
				in *mocks.MockInMemoryRep,
				words *mocks.MockWords,
			) {
				db.EXPECT().
					CreateIndex().
					Return([]core.WordInformation{
						{
							Word: "golang",
							ID:   1,
							Url:  "https://example.com/1.jpg",
						},
						{
							Word: "docker",
							ID:   2,
							Url:  "https://example.com/2.jpg",
						},
					}, nil).
					Times(1)

				in.EXPECT().
					RebuildIndex(map[string][]core.ImageInformation{
						"golang": {
							{
								ID:  1,
								Url: "https://example.com/1.jpg",
							},
						},
						"docker": {
							{
								ID:  2,
								Url: "https://example.com/2.jpg",
							},
						},
					}).
					Return(nil).
					Times(1)
			},
		},
		{
			name: "DB_Create_Error",
			mockBehavior: func(
				db *mocks.MockDB,
				in *mocks.MockInMemoryRep,
				words *mocks.MockWords,
			) {
				db.EXPECT().
					CreateIndex().
					Return(nil, errors.New("db error")).
					Times(1)
			},
			wantErr: true,
		},
		{
			name: "Inmemory_Error",
			mockBehavior: func(
				db *mocks.MockDB,
				in *mocks.MockInMemoryRep,
				words *mocks.MockWords,
			) {
				db.EXPECT().
					CreateIndex().
					Return([]core.WordInformation{
						{
							Word: "golang",
							ID:   1,
							Url:  "https://example.com/1.jpg",
						},
						{
							Word: "docker",
							ID:   2,
							Url:  "https://example.com/2.jpg",
						},
					}, nil).
					Times(1)

				in.EXPECT().
					RebuildIndex(map[string][]core.ImageInformation{
						"golang": {
							{
								ID:  1,
								Url: "https://example.com/1.jpg",
							},
						},
						"docker": {
							{
								ID:  2,
								Url: "https://example.com/2.jpg",
							},
						},
					}).
					Return(errors.New("rebuild error")).
					Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockDB(ctrl)
			in := mocks.NewMockInMemoryRep(ctrl)
			words := mocks.NewMockWords(ctrl)

			tt.mockBehavior(db, in, words)

			service := core.NewService(logger, db, words, in)

			err := service.RebuildIndex()

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestService_DeleteIndex(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mocks.NewMockDB(ctrl)
	in := mocks.NewMockInMemoryRep(ctrl)
	words := mocks.NewMockWords(ctrl)
	s := core.NewService(logger, db, words, in)
	in.EXPECT().DeleteIndex()
	s.DeleteIndex()
}
