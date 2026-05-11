package update

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"yadro.com/course/api/core"
	updatepb "yadro.com/course/proto/update"
)

type fakeUpdateClient struct {
	pingFn   func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	statusFn func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatusReply, error)
	statsFn  func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatsReply, error)
	updateFn func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	dropFn   func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

func (f fakeUpdateClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return f.pingFn(ctx, in, opts...)
}

func (f fakeUpdateClient) Status(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatusReply, error) {
	return f.statusFn(ctx, in, opts...)
}

func (f fakeUpdateClient) Stats(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatsReply, error) {
	return f.statsFn(ctx, in, opts...)
}

func (f fakeUpdateClient) Update(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return f.updateFn(ctx, in, opts...)
}

func (f fakeUpdateClient) Drop(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return f.dropFn(ctx, in, opts...)
}

var logger = slog.New(slog.NewTextHandler(io.Discard, nil))
var fake = fakeUpdateClient{
	pingFn: func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
		return &emptypb.Empty{}, nil
	},
	statusFn: func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatusReply, error) {
		return &updatepb.StatusReply{}, nil
	},
	statsFn: func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatsReply, error) {
		return &updatepb.StatsReply{}, nil
	},
	updateFn: func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
		return &emptypb.Empty{}, nil
	},
	dropFn: func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
		return &emptypb.Empty{}, nil
	},
}

func TestClient_Ping(t *testing.T) {
	t.Parallel()
	t.Run("OK", func(t *testing.T) {
		copy := fake
		c := Client{
			client: copy,
			log:    logger,
		}

		err := c.Ping(context.Background())
		require.NoError(t, err)
	})

	t.Run("ERROR", func(t *testing.T) {
		copy := fake
		copy.pingFn = func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
			return nil, errors.New("ping failed")
		}
		c := Client{
			client: copy,
			log:    logger,
		}

		err := c.Ping(context.Background())
		require.Error(t, err)
	})

}

func TestClient_Status(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		reply      *updatepb.StatusReply
		wantStatus core.UpdateStatus
	}{
		{
			name:       "IDLE",
			reply:      &updatepb.StatusReply{Status: updatepb.Status_STATUS_IDLE},
			wantStatus: core.StatusUpdateIdle,
		},
		{
			name:       "RUNNING",
			reply:      &updatepb.StatusReply{Status: updatepb.Status_STATUS_RUNNING},
			wantStatus: core.StatusUpdateRunning,
		},
		{
			name:       "UNKNOWN",
			reply:      &updatepb.StatusReply{Status: updatepb.Status_STATUS_UNSPECIFIED},
			wantStatus: core.StatusUpdateUnknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			copy := fake
			copy.statusFn = func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatusReply, error) {
				return tc.reply, nil
			}
			c := Client{
				client: copy,
				log:    logger,
			}

			res, err := c.Status(context.Background())
			require.NoError(t, err)
			require.Equal(t, tc.wantStatus, res)
		})
	}
}

func TestClient_Stats(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		copy := fake
		copy.statsFn = func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatsReply, error) {
			return &updatepb.StatsReply{
				WordsTotal:    100,
				WordsUnique:   50,
				ComicsFetched: 10,
				ComicsTotal:   200,
			}, nil
		}
		c := Client{
			client: copy,
			log:    logger,
		}

		res, err := c.Stats(context.Background())
		require.NoError(t, err)
		require.Equal(t, core.UpdateStats{
			WordsTotal:    100,
			WordsUnique:   50,
			ComicsFetched: 10,
			ComicsTotal:   200,
		}, res)
	})

	t.Run("error", func(t *testing.T) {
		copy := fake
		copy.statsFn = func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatsReply, error) {
			return nil, errors.New("stats failed")
		}
		c := Client{
			client: copy,
			log:    logger,
		}

		res, err := c.Stats(context.Background())
		require.Error(t, err)
		require.Equal(t, core.UpdateStats{}, res)
	})
}

func TestClient_Update(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		copy := fake
		c := Client{
			client: copy,
			log:    logger,
		}

		err := c.Update(context.Background())
		require.NoError(t, err)
	})

	t.Run("ALREADY_UPDATING", func(t *testing.T) {
		copy := fake
		copy.updateFn = func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
			return nil, gstatus.Error(codes.Code(alreadyUpdating), "already updating")
		}
		c := Client{
			client: copy,
			log:    logger,
		}

		err := c.Update(context.Background())
		require.ErrorIs(t, err, core.ErrAlreadyUpdating)
	})

	t.Run("ERROR", func(t *testing.T) {
		copy := fake
		copy.updateFn = func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
			return nil, errors.New("update failed")
		}
		c := Client{
			client: copy,
			log:    logger,
		}

		err := c.Update(context.Background())
		require.Error(t, err)
		require.NotErrorIs(t, err, core.ErrAlreadyUpdating)
	})
}

func TestClient_Drop(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		copy := fake
		c := Client{
			client: copy,
			log:    logger,
		}
		err := c.Drop(context.Background())
		require.NoError(t, err)
	})

	t.Run("ERROR", func(t *testing.T) {
		copy := fake
		copy.dropFn = func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
			return nil, errors.New("drop failed")
		}

		c := Client{
			client: copy,
			log:    logger,
		}
		err := c.Drop(context.Background())
		require.Error(t, err)
	})
}
