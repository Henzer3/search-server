package search

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	searchpb "yadro.com/course/proto/search"
)

type Client struct {
	conn   *grpc.ClientConn
	log    *slog.Logger
	client searchpb.SearchClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("creating newUpdateClient error", "err", err)
		return nil, err
	}
	return &Client{
		conn:   conn,
		client: searchpb.NewSearchClient(conn),
		log:    log,
	}, nil
}

func (c Client) Close() error {
	return c.conn.Close()
}

func (c Client) Ping(ctx context.Context) error {
	if _, err := c.client.Ping(ctx, &emptypb.Empty{}); err != nil {
		c.log.Error("Pinging error in updateClient", "err", err)
		return err
	}
	return nil
}

func (c Client) ISearch(ctx context.Context, phrase string, limit int) ([]core.ImageInformation, error) {
	return c.search(ctx, phrase, limit, c.client.ISearch)
}

func (c Client) Search(ctx context.Context, phrase string, limit int) ([]core.ImageInformation, error) {
	return c.search(ctx, phrase, limit, c.client.Search)
}

func (c Client) search(ctx context.Context, phrase string, limit int, search func(ctx context.Context, in *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.SearchReply, error)) ([]core.ImageInformation, error) {
	out, err := search(ctx, &searchpb.SearchRequest{Phrase: phrase, Limit: int64(limit)})
	if err != nil {
		c.log.Error("cant get result of search in client", "err", err)
		return nil, err
	}

	comics := make([]core.ImageInformation, 0, len(out.Images))

	for _, v := range out.Images {
		comics = append(comics, core.ImageInformation{ID: int(v.Id), Url: v.Url})
	}

	return comics, nil
}
