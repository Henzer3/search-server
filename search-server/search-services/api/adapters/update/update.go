package update

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	updatepb "yadro.com/course/proto/update"
)

const alreadyUpdating = 14

type Client struct {
	conn   *grpc.ClientConn
	log    *slog.Logger
	client updatepb.UpdateClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("creating newUpdateClient error", "err", err)
		return nil, err
	}
	return &Client{
		conn:   conn,
		client: updatepb.NewUpdateClient(conn),
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

func (c Client) Status(ctx context.Context) (core.UpdateStatus, error) {
	statusReply, _ := c.client.Status(ctx, &emptypb.Empty{})
	var status core.UpdateStatus
	switch statusReply.Status {
	case updatepb.Status_STATUS_IDLE:
		status = core.StatusUpdateIdle
	case updatepb.Status_STATUS_RUNNING:
		status = core.StatusUpdateRunning
	default:
		status = core.StatusUpdateUnknown
	}

	return status, nil
}

func (c Client) Stats(ctx context.Context) (core.UpdateStats, error) {
	statsReply, err := c.client.Stats(ctx, &emptypb.Empty{})
	if err != nil {
		c.log.Error("getting stats in updateClient", "err", err)
		return core.UpdateStats{}, err
	}
	return core.UpdateStats{WordsTotal: int(statsReply.WordsTotal), WordsUnique: int(statsReply.WordsUnique),
		ComicsFetched: int(statsReply.ComicsFetched), ComicsTotal: int(statsReply.ComicsTotal)}, nil
}

func (c Client) Update(ctx context.Context) error {
	if _, err := c.client.Update(ctx, &emptypb.Empty{}); err != nil {
		if status.Code(err) == alreadyUpdating {
			return core.ErrAlreadyUpdating
		}
		c.log.Error("cant update in client", "err", err)
		return err
	}
	return nil
}

func (c Client) Drop(ctx context.Context) error {
	if _, err := c.client.Drop(ctx, &emptypb.Empty{}); err != nil {
		c.log.Error("droping db in updateClient", "err", err)
		return err
	}
	return nil
}
