package words

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
)

type fakeWordsClient struct {
	norm func(ctx context.Context, in *wordspb.WordsRequest, opts ...grpc.CallOption) (*wordspb.WordsReply, error)
	ping func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

func (f fakeWordsClient) Norm(ctx context.Context, in *wordspb.WordsRequest, opts ...grpc.CallOption) (*wordspb.WordsReply, error) {
	return f.norm(ctx, in, opts...)
}

func (f fakeWordsClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return f.ping(ctx, in, opts...)
}

func TestClient_Norm(t *testing.T) {
	t.Parallel()

	c := Client{
		client: fakeWordsClient{
			norm: func(ctx context.Context, in *wordspb.WordsRequest, opts ...grpc.CallOption) (*wordspb.WordsReply, error) {
				require.Equal(t, in.Phrase, "I love golang")
				return &wordspb.WordsReply{Words: []string{"I", "love", "golang"}}, nil
			},
		},
		log: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	res, err := c.Norm(context.Background(), "I love golang")
	require.NoError(t, err)
	require.Equal(t, []string{"I", "love", "golang"}, res)
}

func TestClient_NormError(t *testing.T) {
	t.Parallel()

	c := Client{
		client: fakeWordsClient{
			norm: func(ctx context.Context, in *wordspb.WordsRequest, opts ...grpc.CallOption) (*wordspb.WordsReply, error) {
				return nil, errors.New("Normalazing error")
			},
		},
		log: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	res, err := c.Norm(context.Background(), "I love golang")
	require.Error(t, err)
	require.Nil(t, res)
}

func TestWordsAdapter_Ping(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		c := Client{
			client: fakeWordsClient{
				ping: func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
					return new(emptypb.Empty), nil
				},
			},
			log: slog.New(slog.NewTextHandler(io.Discard, nil)),
		}

		err := c.Ping(context.Background())
		require.NoError(t, err)
	})

	t.Run("ERROR", func(t *testing.T) {
		c := Client{
			client: fakeWordsClient{
				ping: func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
					return nil, errors.New("ping error")
				},
			},
			log: slog.New(slog.NewTextHandler(io.Discard, nil)),
		}

		err := c.Ping(context.Background())
		require.Error(t, err)
	})

}
