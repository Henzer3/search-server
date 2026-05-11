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

	content := `
log_level: INFO
search_address: localhost:18080
db_address: localhost:18082
words_address: localhost:18081
index_ttl: 12h
broker_adress: nats://localhost:4222
`
	err := os.WriteFile(configPath, []byte(content), 0o644)
	require.NoError(t, err)

	cfg := MustLoad(configPath)

	require.Equal(t, "INFO", cfg.LogLevel)
	require.Equal(t, "localhost:18080", cfg.Address)
	require.Equal(t, "localhost:18082", cfg.DBAddress)
	require.Equal(t, "localhost:18081", cfg.WordsAddress)
	require.Equal(t, 12*time.Hour, cfg.IndexTtl)
	require.Equal(t, "nats://localhost:4222", cfg.NatsAdress)
}

func TestMustLoad_Defaults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	err := os.WriteFile(configPath, []byte("{}\n"), 0o644)
	require.NoError(t, err)

	cfg := MustLoad(configPath)

	require.Equal(t, "DEBUG", cfg.LogLevel)
	require.Equal(t, "localhost:80", cfg.Address)
	require.Equal(t, "localhost:82", cfg.DBAddress)
	require.Equal(t, "localhost:81", cfg.WordsAddress)
	require.Equal(t, 24*time.Hour, cfg.IndexTtl)
	require.Equal(t, "nats://nats:4222", cfg.NatsAdress)
}
