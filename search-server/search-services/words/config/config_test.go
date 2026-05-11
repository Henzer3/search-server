package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad_Success(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte("port: 9090\n")
	err := os.WriteFile(configPath, content, 0o644)
	require.NoError(t, err, "cant write in tempfile")

	cfg, err := Load(configPath)
	require.NoError(t, err, "cant load")

	require.NotEqual(t, nil, cfg)
	require.Equal(t, "9090", cfg.Port)
}

func TestLoad_FileNotFound(t *testing.T) {
	t.Parallel()

	cfg, err := Load("not-existing-file.yaml")
	require.Error(t, err)
	require.Nil(t, cfg)
	require.ErrorContains(t, err, "not-existing-file.yaml")
}
