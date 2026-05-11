package sso

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"yadro.com/course/api/core"
	ssopb "yadro.com/course/proto/sso"
)

type Client struct {
	conn   *grpc.ClientConn
	log    *slog.Logger
	client ssopb.AuthClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("creating newUpdateClient error", "err", err)
		return nil, err
	}
	return &Client{
		conn:   conn,
		client: ssopb.NewAuthClient(conn),
		log:    log,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Login(ctx context.Context, in core.LoginRequest) ([]byte, error) {
	res, err := c.client.Login(ctx, &ssopb.LoginRequest{Email: in.Email, Password: in.Password, AppId: in.AppId})
	if err != nil {
		return nil, err
	}

	return []byte(res.GetToken()), nil
}

func (c *Client) Register(ctx context.Context, email string, password string) (int64, error) {
	res, err := c.client.Register(ctx, &ssopb.RegisterRequest{Email: email, Password: password})
	if err != nil {
		return 0, err
	}

	return res.GetUserId(), nil
}
func (c *Client) Verify(ctx context.Context, token string) (core.UserPermissions, error) {
	res, err := c.client.Verify(ctx, &ssopb.VerifyRequest{Token: token})
	if err != nil {
		return core.UserPermissions{}, err
	}
	return core.UserPermissions{
		ID:      res.GetUserId(),
		Email:   res.GetEmail(),
		AppID:   res.GetAppId(),
		IsAdmin: res.GetIsAdmin(),
	}, nil
}
