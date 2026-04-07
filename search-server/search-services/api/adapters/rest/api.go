package rest

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"sync"

	"github.com/VictoriaMetrics/metrics"
	"yadro.com/course/api/core"
)

type ServerStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type PingResponse struct {
	Replies map[string]string `json:"replies"`
}

type NormResponse struct {
	Words []string `json:"words"`
	Total int      `json:"total"`
}

type StatsResponse struct {
	WordsTotal    int `json:"words_total"`
	WordsUnique   int `json:"words_unique"`
	ComicsFetched int `json:"comics_fetched"`
	ComicsTotal   int `json:"comics_total"`
}

type UpdateStatus struct {
	Status string `json:"status"`
}

type ComicResponse struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type SearchResponse struct {
	Comics []ComicResponse `json:"comics"`
	Total  int             `json:"total"`
}

type UserData struct {
	User     string `json:"name"`
	Password string `json:"password"`
}

func NewWordsHandler(log *slog.Logger, norm core.Normalizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			http.Error(w, core.ErrBadArguments.Error(), http.StatusBadRequest)
			return
		}

		normalizedPhrase, err := norm.Norm(r.Context(), phrase)
		if err != nil {
			if errors.Is(err, core.ErrLimit) {
				http.Error(w, core.ErrBadArguments.Error(), http.StatusBadRequest)
				return
			}
			log.Error("cant normalize", "err", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		answer := NormResponse{Words: normalizedPhrase, Total: len(normalizedPhrase)}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(answer); err != nil {
			log.Error("Internal error during encode json", "err", err)
		}

	}
}

func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		answer := PingResponse{Replies: make(map[string]string)}

		var mu sync.Mutex
		wg := new(sync.WaitGroup)

		for name, v := range pingers {
			wg.Go(func() {
				err := v.Ping(r.Context())
				mu.Lock()
				if err != nil {
					answer.Replies[name] = "unavailable"
				} else {
					answer.Replies[name] = "ok"
				}
				mu.Unlock()
			})
		}

		wg.Wait()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(answer); err != nil {
			log.Error("error during encode json in NewPingHandler", "err", err)
		}
	}
}

func NewUpdateHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := updater.Update(r.Context()); err != nil {
			if errors.Is(err, core.ErrAlreadyUpdating) {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			log.Error("NewUpdateHandler error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func NewUpdateStatsHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := updater.Stats(r.Context())
		if err != nil {
			log.Error("NewUpdateStatsHandler error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(StatsResponse{WordsTotal: stats.WordsTotal, WordsUnique: stats.WordsUnique,
			ComicsFetched: stats.ComicsFetched, ComicsTotal: stats.ComicsTotal}); err != nil {
			log.Error("error during encode json in NewUpdateStatsHandler", "err", err)
		}
	}
}

func NewUpdateStatusHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, _ := updater.Status(r.Context())

		var answer UpdateStatus
		if status == core.StatusUpdateIdle {
			answer.Status = "idle"
		} else {
			answer.Status = "running"
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(answer); err != nil {
			log.Error("error during encode json in NewUpdateStatusHandler", "err", err)
		}
	}
}

func NewDropHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := updater.Drop(r.Context()); err != nil {
			log.Error("NewDropHandler error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func NewSearchHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return newSearchHandler(log, searcher.Search)
}

func NewISearchHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return newSearchHandler(log, searcher.ISearch)
}

func newSearchHandler(log *slog.Logger, searchFunc func(ctx context.Context, phrase string, limit int) ([]core.ImageInformation, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limitStr := r.URL.Query().Get("limit")

		limit := core.DefaultLimitValue
		if limitStr != "" {
			l, err := strconv.Atoi(limitStr)
			if err != nil || l <= 0 {
				http.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}
			limit = l
		}

		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			http.Error(w, "phrase is required", http.StatusBadRequest)
			return
		}

		out, err := searchFunc(r.Context(), phrase, limit)
		if err != nil {
			log.Error("search handler error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		total := len(out)

		response := SearchResponse{
			Comics: make([]ComicResponse, 0, total),
			Total:  total,
		}

		for i := range total {
			response.Comics = append(response.Comics, ComicResponse{ID: out[i].ID, URL: out[i].Url})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("encode json error", "err", err)
		}
	}
}

func NewLoginHandler(log *slog.Logger, auth core.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var userData UserData
		defer func() {
			if err := r.Body.Close(); err != nil {
				log.Error("cant close r.Body in NewLoginHandler", "err", err)
			}
		}()
		if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
			log.Error("cant decode user data", "err", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		token, err := auth.Login(userData.User, userData.Password)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := w.Write(token); err != nil {
			log.Error("cant write in NewLoginHandler", "err", err)
		}
	}
}

func NewMetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics.WritePrometheus(w, true)
	}
}
