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

var logger = slog.New(slog.NewTextHandler(io.Discard, nil))

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

func TestNew_BadAddress(t *testing.T) {
	t.Parallel()

	pub, err := New(logger, "://bad-address")
	require.Error(t, err)
	require.Nil(t, pub)
}

func TestPublisher_Publish(t *testing.T) {
	srv := runTestServer(t)

	pub, err := New(logger, srv.ClientURL())
	require.NoError(t, err)
	require.NotNil(t, pub)
	t.Cleanup(func() {
		pub.Close()
	})

	sub, err := nats.Connect(srv.ClientURL())
	require.NoError(t, err)
	defer sub.Close()

	ch := make(chan *nats.Msg, 1)
	_, err = sub.ChanSubscribe("test.topic", ch)
	require.NoError(t, err)

	err = pub.Publish("test.topic", "event")
	require.NoError(t, err)

	select {
	case msg := <-ch:
		require.Equal(t, "test.topic", msg.Subject)
		require.Equal(t, "event", string(msg.Data))
	case <-time.After(2 * time.Second):
		t.Fatal("message was not received")
	}
}

func TestPublisher_Close(t *testing.T) {
	srv := runTestServer(t)

	pub, err := New(logger, srv.ClientURL())
	require.NoError(t, err)
	require.NotNil(t, pub)

	pub.Close()

	require.True(t, pub.conn.IsClosed())
}
