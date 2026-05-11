package core_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"yadro.com/course/update/core"
	"yadro.com/course/update/core/mocks"
)

var logger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestService(t *testing.T) {
	t.Parallel()

	const (
		Description = "I love golang so much"
		Url         = "https://example.com/1.jpg"
	)

	tests := []struct {
		name      string
		mockSetup func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords, pub *mocks.MockPublisher)
		wantErr   bool
	}{
		{
			name: "OK",
			mockSetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords, pub *mocks.MockPublisher) {
				db.EXPECT().IDs(gomock.Any()).Return([]int{1, 2, 3, 4, 5, 6, 7, 9, 10}, nil).Times(1)
				xkcd.EXPECT().LastID(gomock.Any()).Return(20, nil)

				expectedIDs := map[int]bool{
					8: true, 11: true, 12: true, 13: true, 14: true,
					15: true, 16: true, 17: true, 18: true, 19: true, 20: true,
				}
				xkcd.EXPECT().Get(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, id int) (core.XKCDInfo, error) {
					require.True(t, true, expectedIDs[id], fmt.Sprintf("should add comics with number %d", id))
					return core.XKCDInfo{
						ID:          id,
						URL:         Url,
						Title:       "golang",
						Description: Description,
					}, nil
				}).Times(11)
				words.EXPECT().Norm(gomock.Any(), Description).Return([]string{"golang", "love"}, nil).Times(11)
				db.EXPECT().Add(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, c core.Comics) error {
					require.Equal(t, Url, c.URL)
					require.Equal(t, []string{"golang", "love"}, c.Words)
					require.True(t, expectedIDs[c.ID], "unexpected comic id: %d", c.ID)
					return nil
				}).Times(11)

				pub.EXPECT().Publish("xkcd.db.updated", "XKCD DB has been updated").Return(nil).Times(1)
			},
		},
		{
			name: "IDS_ERROR",
			mockSetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords, pub *mocks.MockPublisher) {
				db.EXPECT().IDs(gomock.Any()).Return(nil, errors.New("Ids error")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "LASTID_ERROR",
			mockSetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords, pub *mocks.MockPublisher) {
				db.EXPECT().IDs(gomock.Any()).Return([]int{1, 2, 3, 4, 5, 6, 7, 9, 10}, nil).Times(1)
				xkcd.EXPECT().LastID(gomock.Any()).Return(0, errors.New("Get last id error"))
			},
			wantErr: true,
		},
		{
			name: "LOAD_DB_ERRORS",
			mockSetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords, pub *mocks.MockPublisher) {
				db.EXPECT().IDs(gomock.Any()).Return([]int{1, 2, 3, 4, 5, 6, 7, 9, 10}, nil).Times(1)
				xkcd.EXPECT().LastID(gomock.Any()).Return(20, nil)

				expectedIDs := map[int]bool{
					8: true, 11: true, 12: true, 13: true, 14: true,
					15: true, 16: true, 17: true, 18: true, 19: true, 20: true,
				}
				xkcd.EXPECT().Get(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, id int) (core.XKCDInfo, error) {
					require.True(t, true, expectedIDs[id], fmt.Sprintf("should add comics with number %d", id))
					if id == 12 {
						return core.XKCDInfo{}, errors.New("cant get info")
					}
					return core.XKCDInfo{
						ID:          id,
						URL:         Url,
						Title:       "golang",
						Description: Description,
					}, nil
				}).Times(11)

				var callCount int32
				words.EXPECT().Norm(gomock.Any(), Description).DoAndReturn(func(ctx context.Context, s string) ([]string, error) {
					if atomic.AddInt32(&callCount, 1) == 1 {
						return nil, errors.New("Normolize error")
					}
					return []string{"golang", "love"}, nil
				}).Times(10)

				db.EXPECT().Add(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, c core.Comics) error {
					require.Equal(t, Url, c.URL)
					require.Equal(t, []string{"golang", "love"}, c.Words)
					require.True(t, expectedIDs[c.ID], "unexpected comic id: %d", c.ID)
					if c.ID == 19 {
						return errors.New("add in db error")
					}
					return nil
				}).Times(9)

				pub.EXPECT().Publish("xkcd.db.updated", "XKCD DB has been updated").Return(nil).Times(1)
			},
		},
		{
			name: "PUBLISH_ERROR",
			mockSetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords, pub *mocks.MockPublisher) {
				db.EXPECT().IDs(gomock.Any()).Return([]int{1, 2, 3, 4, 5, 6, 7, 9, 10}, nil).Times(1)
				xkcd.EXPECT().LastID(gomock.Any()).Return(20, nil)
				xkcd.EXPECT().Get(gomock.Any(), gomock.Any()).Return(core.XKCDInfo{
					ID:          1,
					URL:         Url,
					Title:       "golang",
					Description: Description,
				}, nil).Times(11)
				words.EXPECT().Norm(gomock.Any(), Description).Return([]string{"golang", "love"}, nil).Times(11)
				db.EXPECT().Add(gomock.Any(), gomock.Any()).Return(nil).Times(11)
				pub.EXPECT().Publish("xkcd.db.updated", "XKCD DB has been updated").Return(errors.New("cant publish")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockDB(ctrl)
			xkcd := mocks.NewMockXKCD(ctrl)
			words := mocks.NewMockWords(ctrl)
			pub := mocks.NewMockPublisher(ctrl)

			tc.mockSetup(db, xkcd, words, pub)

			s, err := core.NewService(logger, db, xkcd, words, pub, 5)
			require.NoError(t, err)
			err = s.Update(context.Background())
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestBadArguments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := mocks.NewMockDB(ctrl)
	xkcd := mocks.NewMockXKCD(ctrl)
	words := mocks.NewMockWords(ctrl)
	pub := mocks.NewMockPublisher(ctrl)

	s, err := core.NewService(logger, db, xkcd, words, pub, 0)
	require.Error(t, err)
	require.Nil(t, s)
}

func TestService_Stats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		Mocksetup func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords, pub *mocks.MockPublisher)
		want      core.ServiceStats
		wantErr   bool
	}{
		{
			name: "OK",
			Mocksetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords, pub *mocks.MockPublisher) {
				db.EXPECT().Stats(gomock.Any()).Return(core.DBStats{
					WordsTotal:    100,
					WordsUnique:   40,
					ComicsFetched: 20,
				}, nil).Times(1)

				xkcd.EXPECT().LastID(gomock.Any()).Return(50, nil).Times(1)
			},
			want: core.ServiceStats{
				DBStats: core.DBStats{
					WordsTotal:    100,
					WordsUnique:   40,
					ComicsFetched: 20,
				},
				ComicsTotal: 49,
			},
		},
		{
			name: "DB_STATS_ERROR",
			Mocksetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords, pub *mocks.MockPublisher) {
				db.EXPECT().Stats(gomock.Any()).Return(core.DBStats{}, errors.New("db stats error")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "LASTID_ERROR",
			Mocksetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords, pub *mocks.MockPublisher) {
				db.EXPECT().Stats(gomock.Any()).Return(core.DBStats{
					WordsTotal:    100,
					WordsUnique:   40,
					ComicsFetched: 20,
				}, nil).Times(1)

				xkcd.EXPECT().LastID(gomock.Any()).Return(0, errors.New("last id error")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockDB(ctrl)
			xkcd := mocks.NewMockXKCD(ctrl)
			words := mocks.NewMockWords(ctrl)
			pub := mocks.NewMockPublisher(ctrl)

			tc.Mocksetup(db, xkcd, words, pub)

			s, err := core.NewService(logger, db, xkcd, words, pub, 1)
			require.NoError(t, err)

			got, err := s.Stats(context.Background())
			if tc.wantErr {
				require.Error(t, err)
				require.Equal(t, core.ServiceStats{}, got)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestService_Drop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		Mocksetup func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords, pub *mocks.MockPublisher)
		wantErr   bool
	}{
		{
			name: "OK",
			Mocksetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords, pub *mocks.MockPublisher) {
				db.EXPECT().Drop(gomock.Any()).Return(nil).Times(1)
				pub.EXPECT().Publish("xkcd.db.deleted", "XKCD DB has been deleted").Return(nil).Times(1)
			},
		},
		{
			name: "DB_DROP_ERROR",
			Mocksetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords, pub *mocks.MockPublisher) {
				db.EXPECT().Drop(gomock.Any()).Return(errors.New("drop error")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "PULISH_ERROR",
			Mocksetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords, pub *mocks.MockPublisher) {
				db.EXPECT().Drop(gomock.Any()).Return(nil).Times(1)
				pub.EXPECT().Publish("xkcd.db.deleted", "XKCD DB has been deleted").Return(errors.New("publish error")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockDB(ctrl)
			xkcd := mocks.NewMockXKCD(ctrl)
			words := mocks.NewMockWords(ctrl)
			pub := mocks.NewMockPublisher(ctrl)

			tc.Mocksetup(db, xkcd, words, pub)

			s, err := core.NewService(logger, db, xkcd, words, pub, 1)
			require.NoError(t, err)

			err = s.Drop(context.Background())
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
