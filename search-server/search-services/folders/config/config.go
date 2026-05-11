package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel       string `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	FoldersAddress string `yaml:"folders_address" env:"FOLDERS_ADDRESS" env-required:"true"`
	DBAddress      string `yaml:"db_address" env:"DB_ADDRESS" env-required:"true"`
	SearchAdress   string `yaml:"search_address" env:"SEARCH_ADDRESS" env-required:"true"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
