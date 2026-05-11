package core

import "context"

type UpdateStatus string

const (
	StatusUpdateUnknown UpdateStatus = "unknown"
	StatusUpdateIdle    UpdateStatus = "idle"
	StatusUpdateRunning UpdateStatus = "running"
	DefaultLimitValue                = 10
)

type UpdateStats struct {
	WordsTotal    int
	WordsUnique   int
	ComicsFetched int
	ComicsTotal   int
}

type ImageInformation struct {
	ID  int
	Url string
}

type Folder struct {
	FolderID int64
	Name     string
}

type LoginRequest struct {
	Email    string
	Password string
	AppId    int32
}

type UserPermissions struct {
	ID      int64
	Email   string
	AppID   int32
	IsAdmin bool
}

type userContextKey struct{}

func WithUser(ctx context.Context, user UserPermissions) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

func UserFromContext(ctx context.Context) (UserPermissions, bool) {
	user, ok := ctx.Value(userContextKey{}).(UserPermissions)
	return user, ok
}
