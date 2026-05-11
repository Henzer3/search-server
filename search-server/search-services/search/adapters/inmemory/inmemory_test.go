package inmemory

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"yadro.com/course/search/core"
)

var testLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestIndex_RebuildIndex(t *testing.T) {
	t.Parallel()

	idx := NewRep(testLogger)

	newRep := map[string][]core.ImageInformation{
		"golang": {
			{ID: 1, Url: "https://example.com/1.jpg"},
			{ID: 2, Url: "https://example.com/2.jpg"},
		},
	}

	err := idx.RebuildIndex(newRep)
	require.NoError(t, err)
	require.Equal(t, newRep, idx.rep)
}

func TestIndex_DeleteIndex(t *testing.T) {
	t.Parallel()
	idx := NewRep(testLogger)
	err := idx.RebuildIndex(map[string][]core.ImageInformation{
		"golang": {
			{ID: 1, Url: "https://example.com/1.jpg"},
		},
	})
	require.NoError(t, err)

	idx.DeleteIndex()
	require.Empty(t, idx.rep)
}

func TestIndex_Search(t *testing.T) {
	t.Parallel()

	idx := NewRep(testLogger)

	newRep := map[string][]core.ImageInformation{
		"golang": {
			{ID: 1, Url: "https://example.com/1.jpg"},
			{ID: 2, Url: "https://example.com/2.jpg"},
		},
		"linux": {
			{ID: 1, Url: "https://example.com/1.jpg"},
		},
	}

	err := idx.RebuildIndex(newRep)
	require.NoError(t, err)
	require.Equal(t, newRep, idx.rep)

	res, err := idx.Search(context.Background(), []string{"golang", "linux"})
	require.NoError(t, err)
	require.Equal(t, 2, len(res))
}

func TestSearch(t *testing.T) {
	idx := NewRep(testLogger)

	newRep := map[string][]core.ImageInformation{
		"golang": {
			{ID: 1, Url: "https://example.com/1.jpg"},
			{ID: 2, Url: "https://example.com/2.jpg"},
		},
		"linux": {
			{ID: 1, Url: "https://example.com/1.jpg"},
		},
		"backend": {
			{ID: 2, Url: "https://example.com/2.jpg"},
			{ID: 3, Url: "https://example.com/3.jpg"},
			{ID: 4, Url: "https://example.com/4.jpg"},
		},
		"docker": {
			{ID: 3, Url: "https://example.com/3.jpg"},
		},
		"kubernetes": {
			{ID: 4, Url: "https://example.com/4.jpg"},
			{ID: 5, Url: "https://example.com/5.jpg"},
		},
		"database": {
			{ID: 2, Url: "https://example.com/2.jpg"},
			{ID: 5, Url: "https://example.com/5.jpg"},
			{ID: 6, Url: "https://example.com/6.jpg"},
			{ID: 7, Url: "https://example.com/7.jpg"},
		},
	}
	err := idx.RebuildIndex(newRep)
	require.NoError(t, err)
	require.Equal(t, newRep, idx.rep)

	var tests = []struct {
		name      string
		phrase    []string
		wantCount int
	}{
		{
			name:      "Empty",
			phrase:    []string{},
			wantCount: 0,
		},
		{
			name:      "NoMatches",
			phrase:    []string{"apple", "windows"},
			wantCount: 0,
		},
		{
			name:      "TwoMatches",
			phrase:    []string{"golang", "linux"},
			wantCount: 2,
		},
		{
			name:      "AllMatches",
			phrase:    []string{"golang", "linux", "backend", "docker", "database", "kubernetes"},
			wantCount: 7,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res, err := idx.Search(context.Background(), tc.phrase)
			require.NoError(t, err)
			require.Equal(t, tc.wantCount, len(res))
		})
	}
}
