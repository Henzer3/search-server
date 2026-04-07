package core

import (
	"context"
	"log/slog"
	"sort"
)

type Service struct {
	log         *slog.Logger
	db          DB
	words       Words
	inMemoryRep InMemoryRep
}

func NewService(log *slog.Logger, db DB, words Words, inMemory InMemoryRep) *Service {
	return &Service{log: log, db: db, words: words, inMemoryRep: inMemory}
}

func (s *Service) Search(ctx context.Context, phrase string, limit int) ([]ImageInformation, error) {
	words, err := s.words.Norm(ctx, phrase)
	if err != nil {
		s.log.Error("cant normolize words", "err", err)
		return nil, err
	}
	imagesInfo, err := s.db.Search(ctx, words)
	if err != nil {
		s.log.Error("cant search in db words", "err", err)
		return nil, err
	}

	return imagesInfo[0:min(limit, len(imagesInfo))], nil
}

func (s *Service) ISearch(ctx context.Context, phrase string, limit int) ([]ImageInformation, error) {
	words, err := s.words.Norm(ctx, phrase)
	if err != nil {
		s.log.Error("cant normolize words", "err", err)
		return nil, err
	}
	quantityComics, err := s.inMemoryRep.Search(ctx, words)
	if err != nil {
		s.log.Error("cant Isearch in InMemoryRep", "err", err)
		return nil, err
	}

	sort.Slice(quantityComics, func(i, j int) bool {
		if quantityComics[i].Total == quantityComics[j].Total {
			return quantityComics[i].ImageInfo.ID < quantityComics[j].ImageInfo.ID
		}
		return quantityComics[i].Total > quantityComics[j].Total
	})

	lim := min(limit, len(quantityComics))

	imagesInfo := make([]ImageInformation, 0, lim)

	for _, v := range quantityComics[0:lim] {
		imagesInfo = append(imagesInfo, v.ImageInfo)
	}

	return imagesInfo, nil
}

func (s *Service) RebuildIndex() error {
	wordsInfo, err := s.db.CreateIndex()
	if err != nil {
		s.log.Error("cant rebuild index", "err", err)
		return err
	}

	rep := make(map[string][]ImageInformation)

	for _, v := range wordsInfo {
		rep[v.Word] = append(rep[v.Word], ImageInformation{ID: v.ID, Url: v.Url})
	}

	if err := s.inMemoryRep.RebuildIndex(rep); err != nil {
		s.log.Error("cant rebuild index", "err", err)
		return err
	}
	return nil
}

func (s *Service) DeleteIndex() {
	s.inMemoryRep.DeleteIndex()
}
