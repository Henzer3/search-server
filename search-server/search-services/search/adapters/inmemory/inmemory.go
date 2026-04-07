package inmemory

import (
	"context"
	"log/slog"
	"sync"

	"yadro.com/course/search/core"
)

type Index struct {
	log *slog.Logger
	rep map[string][]core.ImageInformation
	mu  sync.RWMutex
}

func NewRep(log *slog.Logger) *Index {
	return &Index{log: log, rep: make(map[string][]core.ImageInformation)}
}

func (i *Index) RebuildIndex(newRep map[string][]core.ImageInformation) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.rep = newRep
	return nil
}

func (i *Index) DeleteIndex() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.rep = make(map[string][]core.ImageInformation)
}

func (i *Index) Search(_ context.Context, words []string) ([]core.QuantityComics, error) {
	mapOut := make(map[int]core.QuantityComics)

	// специально mu.Lock не внутри цикла, так как обновление мапы происходит редко
	i.mu.RLock()
	for _, word := range words {
		comics, ok := i.rep[word]
		if ok {
			for _, v := range comics {
				item := mapOut[v.ID]
				item.ImageInfo = v
				item.Total++
				mapOut[v.ID] = item
			}
		}
	}
	i.mu.RUnlock()

	comics := make([]core.QuantityComics, 0, len(mapOut))

	for _, v := range mapOut {
		comics = append(comics, v)
	}

	return comics, nil

}
