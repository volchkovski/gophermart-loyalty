package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Addr        string `env:"RUN_ADDRESS"`
	DSN         string `env:"DATABASE_URI"`
	AccrualAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	Secret      string `env:"AUTH_SECRET"`
	LogLevel    string `env:"LOG_LEVEL"`
	Env         string `env:"ENVIRONMENT"`
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
	flag.StringVar(&cfg.DSN, "d", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable", "postgres data source name")
	flag.StringVar(&cfg.AccrualAddr, "r", "localhost:8081", "host and port to run accrual system")
	flag.StringVar(&cfg.Secret, "s", "test", "secret for jwt generation")
	flag.StringVar(&cfg.LogLevel, "l", "debug", "level of logging")
	flag.StringVar(&cfg.Env, "e", "local", "environment: prod, local")
	flag.Parse()
}
