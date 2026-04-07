package core

import "context"

type Searcher interface {
	Search(ctx context.Context, phrase string, limit int) ([]ImageInformation, error)
	ISearch(ctx context.Context, phrase string, limit int) ([]ImageInformation, error)
}

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}

type DB interface {
	Search(ctx context.Context, words []string) ([]ImageInformation, error)
	CreateIndex() ([]WordInformation, error)
}

type InMemoryRep interface {
	Search(ctx context.Context, words []string) ([]QuantityComics, error)
	RebuildIndex(rep map[string][]ImageInformation) error
	DeleteIndex()
}
