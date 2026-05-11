package core

import "context"

//go:generate go run go.uber.org/mock/mockgen -source=ports.go -destination=mocks/mock.go -package=mocks
type Searcher interface {
	Search(ctx context.Context, phrase string, limit int) ([]ImageInformation, error)
	ISearch(ctx context.Context, phrase string, limit int) ([]ImageInformation, error)
	GetComics(ctx context.Context, comicsID int) (string, error)
}

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}

type DB interface {
	Search(ctx context.Context, words []string) ([]ImageInformation, error)
	CreateIndex() ([]WordInformation, error)
	GetComics(ctx context.Context, comicsID int) (ImageInformation, error)
}

type InMemoryRep interface {
	Search(ctx context.Context, words []string) ([]QuantityComics, error)
	RebuildIndex(rep map[string][]ImageInformation) error
	DeleteIndex()
}
