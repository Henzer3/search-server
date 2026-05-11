package xkcd

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"yadro.com/course/update/core"
)

var logger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestNewClient(t *testing.T) {
	t.Parallel()

	t.Run("EmptyUrl", func(t *testing.T) {
		c, err := NewClient("", time.Second, logger)
		require.Error(t, err)
		require.Nil(t, c)
	})

	t.Run("Success", func(t *testing.T) {
		c, err := NewClient("http://example.com", time.Second, logger)
		require.NoError(t, err)
		require.NotNil(t, c)
	})
}

func TestClient_Get(t *testing.T) {
	t.Parallel()

	t.Run("BadId", func(t *testing.T) {
		c, err := NewClient("http://example.com", time.Second, logger)
		require.NoError(t, err)

		res, err := c.Get(context.Background(), 0)
		require.Error(t, err)
		require.Equal(t, core.ErrBadArguments, err)
		require.Equal(t, core.XKCDInfo{}, res)
	})

	t.Run("Success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/123/info.0.json", r.URL.Path)

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"num": 123,
				"safe_title": "Safe title",
				"transcript": "Transcript text",
				"alt": "Alt text",
				"img": "https://imgs.xkcd.com/comics/test.png",
				"title": "Title text"
			}`))
		}))
		defer srv.Close()

		c, err := NewClient(srv.URL, time.Second, logger)
		require.NoError(t, err)

		res, err := c.Get(context.Background(), 123)
		require.NoError(t, err)
		require.Equal(t, core.XKCDInfo{
			ID:          123,
			URL:         "https://imgs.xkcd.com/comics/test.png",
			Title:       "Title text",
			Description: "Safe title Title text Alt text Transcript text",
		}, res)
	})

	t.Run("Not_200_Status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		c, err := NewClient(srv.URL, time.Second, logger)
		require.NoError(t, err)

		res, err := c.Get(context.Background(), 123)
		require.Error(t, err)
		require.Equal(t, core.XKCDInfo{}, res)
	})

	t.Run("BadJson", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{bad json}`))
		}))
		defer srv.Close()

		c, err := NewClient(srv.URL, time.Second, logger)
		require.NoError(t, err)

		res, err := c.Get(context.Background(), 123)
		require.Error(t, err)
		require.Equal(t, core.XKCDInfo{}, res)
	})
}

func TestClient_LastID(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/info.0.json", r.URL.Path)

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"num": 321}`))
		}))
		defer srv.Close()

		c, err := NewClient(srv.URL, time.Second, logger)
		require.NoError(t, err)

		id, err := c.LastID(context.Background())
		require.NoError(t, err)
		require.Equal(t, 321, id)
	})

	t.Run("Not_200_Status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		c, err := NewClient(srv.URL, time.Second, logger)
		require.NoError(t, err)

		id, err := c.LastID(context.Background())
		require.Error(t, err)
		require.Equal(t, 0, id)
	})

	t.Run("BadJson", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{bad json}`))
		}))
		defer srv.Close()

		c, err := NewClient(srv.URL, time.Second, logger)
		require.NoError(t, err)

		id, err := c.LastID(context.Background())
		require.Error(t, err)
		require.Equal(t, 0, id)
	})
}
