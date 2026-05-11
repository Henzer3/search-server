package search

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	searchpb "yadro.com/course/proto/search"
)

type fakeSearchClient struct {
	pingFn      func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	searchFn    func(ctx context.Context, in *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.SearchReply, error)
	isearchFn   func(ctx context.Context, in *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.SearchReply, error)
	getcomicsFn func(ctx context.Context, in *searchpb.GetComicsRequest, opts ...grpc.CallOption) (*searchpb.GetComicsResponse, error)
}

func (f fakeSearchClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return f.pingFn(ctx, in, opts...)
}

func (f fakeSearchClient) Search(ctx context.Context, in *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.SearchReply, error) {
	return f.searchFn(ctx, in, opts...)
}

func (f fakeSearchClient) ISearch(ctx context.Context, in *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.SearchReply, error) {
	return f.isearchFn(ctx, in, opts...)
}

func (f fakeSearchClient) GetComics(ctx context.Context, in *searchpb.GetComicsRequest, opts ...grpc.CallOption) (*searchpb.GetComicsResponse, error) {
	return f.getcomicsFn(ctx, in, opts...)
}

var logger = slog.New(slog.NewTextHandler(io.Discard, nil))

var fake = fakeSearchClient{
	pingFn: func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
		return &emptypb.Empty{}, nil
	},
	searchFn: func(ctx context.Context, in *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.SearchReply, error) {
		return &searchpb.SearchReply{}, nil
	},
	isearchFn: func(ctx context.Context, in *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.SearchReply, error) {
		return &searchpb.SearchReply{}, nil
	},
}

func TestClient_Ping(t *testing.T) {
	t.Parallel()
	t.Run("OK", func(t *testing.T) {
		c := Client{
			client: fake,
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

func TestClient_Search(t *testing.T) {
	t.Parallel()

	t.Run("Ok", func(t *testing.T) {
		copy := fake
		copy.searchFn = func(ctx context.Context, in *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.SearchReply, error) {
			require.Equal(t, "golang", in.Phrase)
			require.Equal(t, int64(5), in.Limit)

			return &searchpb.SearchReply{
				Images: []*searchpb.Image{
					{Id: 1, Url: "https://example.com/1.jpg"},
					{Id: 2, Url: "https://example.com/2.jpg"},
				},
			}, nil
		}
		c := Client{
			client: copy,
			log:    logger,
		}

		got, err := c.Search(context.Background(), "golang", 5)
		require.NoError(t, err)
		require.Equal(t, []core.ImageInformation{
			{ID: 1, Url: "https://example.com/1.jpg"},
			{ID: 2, Url: "https://example.com/2.jpg"},
		}, got)
	})

	t.Run("ERROR", func(t *testing.T) {
		copy := fake
		copy.searchFn = func(ctx context.Context, in *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.SearchReply, error) {
			return nil, errors.New("search failed")
		}

		c := Client{
			client: copy,
			log:    logger,
		}

		got, err := c.Search(context.Background(), "golang", 5)
		require.Error(t, err)
		require.Nil(t, got)
	})
}

func TestClient_ISearch(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		copy := fake
		copy.isearchFn = func(ctx context.Context, in *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.SearchReply, error) {
			require.Equal(t, "docker", in.Phrase)
			require.Equal(t, int64(3), in.Limit)

			return &searchpb.SearchReply{
				Images: []*searchpb.Image{
					{Id: 10, Url: "https://example.com/10.jpg"},
					{Id: 20, Url: "https://example.com/20.jpg"},
				},
			}, nil
		}
		c := Client{
			client: copy,
			log:    logger,
		}

		got, err := c.ISearch(context.Background(), "docker", 3)
		require.NoError(t, err)
		require.Equal(t, []core.ImageInformation{
			{ID: 10, Url: "https://example.com/10.jpg"},
			{ID: 20, Url: "https://example.com/20.jpg"},
		}, got)
	})

	t.Run("error", func(t *testing.T) {
		copy := fake
		copy.isearchFn = func(ctx context.Context, in *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.SearchReply, error) {
			return nil, errors.New("isearch failed")
		}

		c := Client{
			client: copy,
			log:    logger,
		}

		got, err := c.ISearch(context.Background(), "docker", 3)
		require.Error(t, err)
		require.Nil(t, got)
	})
}
