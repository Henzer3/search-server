package search

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"yadro.com/course/folders/entity"
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

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetComics(ctx context.Context, comicsID int64) (entity.Comics, error) {
	res, err := c.client.GetComics(ctx, &searchpb.GetComicsRequest{ComicsId: comicsID})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			c.log.Debug("hereee in folders adapter", "err", err)
			return entity.Comics{}, entity.ErrComicsNotExist
		}
		c.log.Error("cant GetComics in search Client")
		return entity.Comics{}, err
	}
	return entity.Comics{ComicsID: res.ComicsId, URL: res.Url}, nil
}
