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
search_concurrency: 8
search_rate: 15
words_address: words-service:81
update_address: update-service:82
search_address: search-service:83
token_ttl: 12h
api_server:
  address: 0.0.0.0:8080
  timeout: 10s
`
	err := os.WriteFile(configPath, []byte(content), 0o644)
	require.NoError(t, err)

	cfg := MustLoad(configPath)

	require.Equal(t, "INFO", cfg.LogLevel)
	require.Equal(t, 8, cfg.SearchConcurrency)
	require.Equal(t, 15, cfg.SearchRate)
	require.Equal(t, "words-service:81", cfg.WordsAddress)
	require.Equal(t, "update-service:82", cfg.UpdateAddress)
	require.Equal(t, "search-service:83", cfg.SearchAddress)
	require.Equal(t, 12*time.Hour, cfg.TokenTTL)

	require.Equal(t, "0.0.0.0:8080", cfg.HTTPConfig.Address)
	require.Equal(t, 10*time.Second, cfg.HTTPConfig.Timeout)
}

func TestMustLoad_Defaults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	err := os.WriteFile(configPath, []byte("{}"), 0o644)
	require.NoError(t, err)

	cfg := MustLoad(configPath)

	require.Equal(t, "DEBUG", cfg.LogLevel)
	require.Equal(t, 1, cfg.SearchConcurrency)
	require.Equal(t, 1, cfg.SearchRate)
	require.Equal(t, "words:81", cfg.WordsAddress)
	require.Equal(t, "update:82", cfg.UpdateAddress)
	require.Equal(t, "search:83", cfg.SearchAddress)
	require.Equal(t, 24*time.Hour, cfg.TokenTTL)

	require.Equal(t, "localhost:80", cfg.HTTPConfig.Address)
	require.Equal(t, 5*time.Second, cfg.HTTPConfig.Timeout)
}
