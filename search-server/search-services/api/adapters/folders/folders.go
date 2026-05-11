package folders

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"yadro.com/course/api/core"
	folderpb "yadro.com/course/proto/folders"
)

type Client struct {
	conn   *grpc.ClientConn
	log    *slog.Logger
	client folderpb.FoldersClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("creating newUpdateClient error", "err", err)
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: folderpb.NewFoldersClient(conn),
		log:    log,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) CreateFolder(ctx context.Context, uid int64, name string) (int64, error) {
	res, err := c.client.CreateFolder(ctx, &folderpb.CreateFolderRequest{UserId: uid, Name: name})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return 0, core.ErrAlreadyExists
		}
		c.log.Error("cant CreateFolder in Client in api", "err", err)
		return 0, err
	}
	return res.GetFolderId(), nil
}

func (c *Client) DeleteFolder(ctx context.Context, uid int64, fid int64) error {
	if _, err := c.client.DeleteFolder(ctx, &folderpb.DeleteFolderRequest{UserId: uid, FolderId: fid}); err != nil {
		if status.Code(err) == codes.FailedPrecondition {
			return core.ErrNotFound
		} else if status.Code(err) == codes.PermissionDenied {
			return core.ErrNoPermissions
		}
		c.log.Error("cant DeletFolder in clietn api", "err", err)
		return err
	}
	return nil
}

func (c *Client) AddComics(ctx context.Context, uid int64, fid int64, cid int64) error {
	if _, err := c.client.AddComics(ctx, &folderpb.AddComicsRequest{UserId: uid, FolderId: fid, ComicsId: cid}); err != nil {
		switch {
		case status.Code(err) == codes.FailedPrecondition:
			return core.ErrNotFound
		case status.Code(err) == codes.PermissionDenied:
			return core.ErrNoPermissions
		case status.Code(err) == codes.AlreadyExists:
			return core.ErrAlreadyExists
		}
		c.log.Error("cant AddComics in client api", "err", err)
		return err
	}

	return nil
}

func (c *Client) DeleteComics(ctx context.Context, uid int64, fid int64, cid int64) error {
	if _, err := c.client.DeleteComics(ctx, &folderpb.DeleteComicsRequest{UserId: uid, FolderId: fid, ComicsId: cid}); err != nil {
		if status.Code(err) == codes.FailedPrecondition {
			return core.ErrNotFound
		} else if status.Code(err) == codes.PermissionDenied {
			return core.ErrNoPermissions
		}
		c.log.Error("cant DeleteComics in client api", "err", err)
		return err
	}

	return nil
}

func (c *Client) ListComics(ctx context.Context, uid int64, fid int64) ([]core.ImageInformation, error) {
	res, err := c.client.ListComics(ctx, &folderpb.ListComicsRequest{UserId: uid, FolderId: fid})
	if err != nil {
		if status.Code(err) == codes.FailedPrecondition {
			return nil, core.ErrNotFound
		} else if status.Code(err) == codes.PermissionDenied {
			return nil, core.ErrNoPermissions
		}
		c.log.Error("cant ListComics in client api", "err", err)
		return nil, err
	}

	ans := make([]core.ImageInformation, 0, len(res.Comics))
	for _, v := range res.Comics {
		ans = append(ans, core.ImageInformation{ID: int(v.GetComicsId()), Url: v.GetUrl()})
	}
	return ans, nil
}

func (c *Client) ListFolders(ctx context.Context, uid int64) ([]core.Folder, error) {
	res, err := c.client.ListFolders(ctx, &folderpb.ListFoldersRequest{UserId: uid})
	if err != nil {
		c.log.Error("cant ListFolders in client api", "err", err)
		return nil, err
	}

	ans := make([]core.Folder, 0, len(res.Folders))
	for _, v := range res.Folders {
		ans = append(ans, core.Folder{FolderID: v.GetFolderId(), Name: v.GetName()})
	}
	return ans, nil
}
