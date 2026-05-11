package nats

import (
	"io"
	"log/slog"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

func runTestServer(t *testing.T) *natsserver.Server {
	t.Helper()

	opts := &natsserver.Options{
		Host: "127.0.0.1",
		Port: -1,
	}

	srv, err := natsserver.NewServer(opts)
	require.NoError(t, err)

	go srv.Start()
	if !srv.ReadyForConnections(5 * time.Second) {
		t.Fatal("nats server not ready")
	}

	t.Cleanup(func() {
		srv.Shutdown()
	})

	return srv
}

func TestSubscriber(t *testing.T) {
	srv := runTestServer(t)
	sub, err := New(slog.New(slog.NewTextHandler(io.Discard, nil)), srv.ClientURL(), 10)
	require.NoError(t, err)

	ch := make(chan struct{}, 3)
	err = sub.Subscribe("test", func(msg *nats.Msg) {
		ch <- struct{}{}
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(sub.subscribers))

	sub.StartListen()

	conn, err := nats.Connect(srv.ClientURL())
	require.NoError(t, err)
	defer conn.Close()

	err = conn.Publish("test", []byte("test has been published"))
	require.NoError(t, err)
	time.Sleep(2 * time.Second)

	err = conn.Publish("test", []byte("test has been published"))
	require.NoError(t, err)
	time.Sleep(2 * time.Second)

	err = conn.Publish("test", []byte("test has been published"))
	require.NoError(t, err)

	err = conn.Flush()
	require.NoError(t, err)

	count := 0
	for range 3 {
		select {
		case <-ch:
			count++
		case <-time.After(5 * time.Second):
		}
	}
	require.Equal(t, 3, count)
	err = sub.Stop()
	require.NoError(t, err)
	require.Empty(t, sub.subscribers)
}
