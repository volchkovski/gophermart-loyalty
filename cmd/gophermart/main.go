package main

import (
	"github.com/volchkovski/gophermart-loyalty/internal/app"
	"github.com/volchkovski/gophermart-loyalty/internal/config"
	"log"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	app.MustRun(cfg)
}
