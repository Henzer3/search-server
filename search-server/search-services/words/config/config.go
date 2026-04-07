package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port string `env:"WORDS_GRPC_PORT" env-default:"8080" yaml:"port"`
}

func Load(configPath string) (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
