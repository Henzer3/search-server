package words

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
)

type Client struct {
	conn   *grpc.ClientConn
	log    *slog.Logger
	client wordspb.WordsClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("creating newClientWords error", "err", err)
		return nil, err
	}

	return &Client{
		conn:   conn,
		log:    log,
		client: wordspb.NewWordsClient(conn),
	}, nil

}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c Client) Norm(ctx context.Context, phrase string) ([]string, error) {
	response, err := c.client.Norm(ctx, &wordspb.WordsRequest{Phrase: phrase})
	if err != nil {
		c.log.Error("Normalazing error", "err", err)
		return nil, err
	}
	return response.Words, nil
}

func (c Client) Ping(ctx context.Context) error {
	if _, err := c.client.Ping(ctx, new(emptypb.Empty)); err != nil {
		c.log.Error("Ping error in WordsClient", "err", err)
		return err
	}
	return nil
}
