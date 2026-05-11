package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMustLoad_Success(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := `log_level: INFO
update_address: localhost:18080
db_address: localhost:18082
words_address: localhost:18081
broker_adress: nats://localhost:4222
xkcd:
  url: xkcd.test
  concurrency: 5
  timeout: 30s
  check_period: 2h
`

	err := os.WriteFile(configPath, []byte(content), 0o644)
	require.NoError(t, err)

	cfg := MustLoad(configPath)

	require.Equal(t, "INFO", cfg.LogLevel)
	require.Equal(t, "localhost:18080", cfg.Address)
	require.Equal(t, "localhost:18082", cfg.DBAddress)
	require.Equal(t, "localhost:18081", cfg.WordsAddress)
	require.Equal(t, "nats://localhost:4222", cfg.NatsAdress)

	require.Equal(t, "xkcd.test", cfg.XKCD.URL)
	require.Equal(t, 5, cfg.XKCD.Concurrency)
	require.Equal(t, 30*time.Second, cfg.XKCD.Timeout)
	require.Equal(t, 2*time.Hour, cfg.XKCD.CheckPeriod)
}

func TestMustLoad_Defaults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	err := os.WriteFile(configPath, []byte("{}"), 0o644)
	require.NoError(t, err)

	cfg := MustLoad(configPath)

	require.Equal(t, "DEBUG", cfg.LogLevel)
	require.Equal(t, "localhost:80", cfg.Address)
	require.Equal(t, "localhost:82", cfg.DBAddress)
	require.Equal(t, "localhost:81", cfg.WordsAddress)
	require.Equal(t, "nats://nats:4222", cfg.NatsAdress)

	require.Equal(t, "xkcd.com", cfg.XKCD.URL)
	require.Equal(t, 1, cfg.XKCD.Concurrency)
	require.Equal(t, 15*time.Second, cfg.XKCD.Timeout)
	require.Equal(t, time.Hour, cfg.XKCD.CheckPeriod)
}
