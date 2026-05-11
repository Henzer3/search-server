package grpc

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	updatepb "yadro.com/course/proto/update"
	"yadro.com/course/update/core"
	"yadro.com/course/update/core/mocks"
)

var logger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestServer_Ping(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockUpdater(ctrl)
	s := NewServer(logger, service)

	res, err := s.Ping(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestServer_Status(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		coreStatus core.ServiceStatus
		wantStatus updatepb.Status
	}{
		{
			name:       "running",
			coreStatus: core.StatusRunning,
			wantStatus: updatepb.Status_STATUS_RUNNING,
		},
		{
			name:       "idle",
			coreStatus: core.StatusIdle,
			wantStatus: updatepb.Status_STATUS_IDLE,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mocks.NewMockUpdater(ctrl)
			service.EXPECT().Status(gomock.Any()).Return(tc.coreStatus).Times(1)

			s := NewServer(logger, service)

			res, err := s.Status(context.Background(), &emptypb.Empty{})
			require.NoError(t, err)
			require.NotNil(t, res)
			require.Equal(t, tc.wantStatus, res.Status)
		})
	}
}

func TestServer_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		updateErr error
		wantErr   bool
		wantCode  codes.Code
		wantMsg   string
	}{
		{
			name:      "Success",
			updateErr: nil,
		},
		{
			name:      "Already_Updating",
			updateErr: core.ErrAlreadyUpdating,
			wantErr:   true,
			wantCode:  codes.Unavailable,
			wantMsg:   "update already in progress",
		},
		{
			name:      "Internal_Error",
			updateErr: errors.New("error"),
			wantErr:   true,
			wantCode:  codes.Internal,
			wantMsg:   "processing failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mocks.NewMockUpdater(ctrl)
			service.EXPECT().Update(gomock.Any()).Return(tc.updateErr).Times(1)

			s := NewServer(logger, service)

			res, err := s.Update(context.Background(), &emptypb.Empty{})

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, res)

				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tc.wantCode, st.Code())
				require.Equal(t, tc.wantMsg, st.Message())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, res)
		})
	}
}

func TestServer_Stats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		stats     core.ServiceStats
		statsErr  error
		wantErr   bool
		wantCode  codes.Code
		wantReply *updatepb.StatsReply
	}{
		{
			name: "Success",
			stats: core.ServiceStats{
				DBStats: core.DBStats{
					WordsTotal:    100,
					WordsUnique:   50,
					ComicsFetched: 10,
				},
				ComicsTotal: 200,
			},
			wantReply: &updatepb.StatsReply{
				WordsTotal:    100,
				WordsUnique:   50,
				ComicsFetched: 10,
				ComicsTotal:   200,
			},
		},
		{
			name:     "Internal_Error",
			statsErr: errors.New("stats failed"),
			wantErr:  true,
			wantCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mocks.NewMockUpdater(ctrl)
			service.EXPECT().Stats(gomock.Any()).Return(tc.stats, tc.statsErr).Times(1)

			s := NewServer(logger, service)

			res, err := s.Stats(context.Background(), &emptypb.Empty{})

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, res)

				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tc.wantCode, st.Code())
				require.Equal(t, "processing failed", st.Message())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantReply, res)
		})
	}
}

func TestServer_Drop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		dropErr  error
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name:    "Success",
			dropErr: nil,
		},
		{
			name:     "Internal error",
			dropErr:  errors.New("drop failed"),
			wantErr:  true,
			wantCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mocks.NewMockUpdater(ctrl)
			service.EXPECT().Drop(gomock.Any()).Return(tc.dropErr).Times(1)

			s := NewServer(logger, service)

			res, err := s.Drop(context.Background(), &emptypb.Empty{})

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, res)

				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tc.wantCode, st.Code())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, res)
		})
	}
}
