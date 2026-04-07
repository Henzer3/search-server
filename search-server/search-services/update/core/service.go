package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
)

type Service struct {
	log         *slog.Logger
	db          DB
	xkcd        XKCD
	words       Words
	publisher   Publisher
	concurrency int
	isUpdating  atomic.Bool
}

const comics404 = 404

func NewService(log *slog.Logger, db DB, xkcd XKCD, words Words, publisher Publisher, concurrency int) (*Service, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("wrong concurrency specified: %d", concurrency)
	}
	return &Service{
		log:         log,
		db:          db,
		xkcd:        xkcd,
		words:       words,
		publisher:   publisher,
		concurrency: concurrency,
	}, nil
}

func (s *Service) Update(ctx context.Context) error {
	if !s.isUpdating.CompareAndSwap(false, true) {
		return ErrAlreadyUpdating
	}
	defer s.isUpdating.Store(false)

	ids, err := s.db.IDs(ctx)
	if err != nil {
		return err
	}

	result := make(map[int]bool, len(ids))
	for _, id := range ids {
		result[id] = true
	}
	result[comics404] = true

	lastId, err := s.xkcd.LastID(ctx)
	if err != nil {
		return err
	}

	ch := make(chan int)
	wg := new(sync.WaitGroup)

	wg.Go(func() {
		defer close(ch)
		for i := 1; i <= lastId; i++ {
			if v := result[i]; !v {
				select {
				case <-ctx.Done():
					return
				case ch <- i:
				}
			}
		}
	})

	for range s.concurrency {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-ch:
					if !ok {
						return
					}
					err := s.loadInDB(ctx, v)
					if err != nil {
						s.log.Error("cant load in db", "err", err, "num", v)
					}
				}
			}
		})
	}
	wg.Wait()
	if err := s.publisher.Publish("xkcd.db.updated", "XKCD DB has been updated"); err != nil {
		s.log.Error("cant publish xkcd.db.updated topic", "err", err)
		return err
	}
	return nil
}

func (s *Service) loadInDB(ctx context.Context, num int) error {
	info, err := s.xkcd.Get(ctx, num)
	if err != nil {
		s.log.Error("cant load info about comics from xkcd", "err", err, "num", num)
		return err
	}
	words, err := s.normalizeDescription(ctx, info.Description)
	if err != nil {
		s.log.Error("cant normolize", "err", err)
		return err
	}

	err = s.db.Add(ctx, Comics{ID: info.ID, URL: info.URL, Words: words})
	if err != nil {
		s.log.Error("cant add in db", "err", err)
		return err
	}
	return nil
}

func (s *Service) normalizeDescription(ctx context.Context, description string) ([]string, error) {
	const maxLen = 4096

	parts := splitByBytes(description, maxLen)
	allWords := make([]string, 0)

	for _, part := range parts {
		words, err := s.words.Norm(ctx, part)
		if err != nil {
			s.log.Error("cant normalize", "err", err)
			return nil, err
		}
		allWords = append(allWords, words...)
	}

	return allWords, nil
}

func splitByBytes(s string, maxLen int) []string {
	bytes := []byte(s)

	parts := make([]string, 0, (len(bytes)+maxLen-1)/maxLen)
	for i := 0; i < len(bytes); i += maxLen {
		end := i + maxLen
		if end > len(bytes) {
			end = len(bytes)
		}
		parts = append(parts, string(bytes[i:end]))
	}

	return parts
}

func (s *Service) Stats(ctx context.Context) (ServiceStats, error) {
	dbStats, err := s.db.Stats(ctx)
	if err != nil {
		s.log.Error("cant get stats from db", "err", err)
		return ServiceStats{}, err
	}

	comicsTotal, err := s.xkcd.LastID(ctx)
	if err != nil {
		s.log.Error("cant get lastid from xkcd", "err", err)
		return ServiceStats{}, err
	}
	// without comics 404
	comicsTotal--
	return ServiceStats{DBStats: dbStats, ComicsTotal: comicsTotal}, nil

}

func (s *Service) Status(ctx context.Context) ServiceStatus {
	if s.isUpdating.Load() {
		return StatusRunning
	}
	return StatusIdle
}

func (s *Service) Drop(ctx context.Context) error {
	if err := s.db.Drop(ctx); err != nil {
		s.log.Error("cant drop db", "err", err)
		return err
	}
	if err := s.publisher.Publish("xkcd.db.deleted", "XKCD DB has been deleted"); err != nil {
		s.log.Error("cant publish xkcd.db.deleted topic", "err", err)
		return err
	}
	return nil
}
