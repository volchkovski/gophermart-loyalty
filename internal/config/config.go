package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	Addr     string `env:"ADDRESS"`
	LogLevel string `env:"LOG_LEVEL"`
	Env      string `env:"ENVIRONMENT"`
}

func New() (*Config, error) {
	cfg := new(Config)
	parseFlags(cfg)
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}
	return cfg, nil
}

func parseFlags(cfg *Config) {
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "host and port to run the app")
	flag.StringVar(&cfg.Addr, "l", "debug", "level of logging")
	flag.StringVar(&cfg.Addr, "e", "local", "environment: prod, local")
	flag.Parse()
}
