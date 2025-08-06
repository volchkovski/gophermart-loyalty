package main

import (
	"github.com/volchkovski/gophermart-loyalty/internal/app"
	"github.com/volchkovski/gophermart-loyalty/internal/config"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}
	app.Run(cfg)
}
