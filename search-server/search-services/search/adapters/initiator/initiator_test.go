package initiator

import (
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInitiator_Success(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	count := 0
	f := func() error {
		mu.Lock()
		defer mu.Unlock()
		count++
		return nil
	}
	initiator, err := NewInitiator(slog.New(slog.NewTextHandler(io.Discard, nil)), 1*time.Second, f)
	require.NoError(t, err)
	time.Sleep(5 * time.Second)
	initiator.Stop()
	mu.Lock()
	c := count
	mu.Unlock()
	require.LessOrEqual(t, 5, c, "should be more than 5")
}

func TestInitiator_BadArguments(t *testing.T) {
	t.Parallel()
	initiator, err := NewInitiator(slog.New(slog.NewTextHandler(io.Discard, nil)), -1*time.Second, func() error {
		return nil
	})
	require.Error(t, err)
	require.Nil(t, initiator)
}
