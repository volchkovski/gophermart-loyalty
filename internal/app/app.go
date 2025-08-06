package app

import (
	"github.com/volchkovski/gophermart-loyalty/internal/config"
	"github.com/volchkovski/gophermart-loyalty/internal/httpserver"
	"github.com/volchkovski/gophermart-loyalty/internal/logger"
	"github.com/volchkovski/gophermart-loyalty/internal/router"
	"os"
	"os/signal"
	"syscall"
)

func Run(cfg *config.Config) {
	if err := logger.Initialize(cfg.LogLevel, cfg.Env); err != nil {
		panic(err)
	}
	r := router.NewHTTPRouter()
	server := httpserver.New(cfg.Addr, r)
	server.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case <-server.Notify():
		return
	case s := <-interrupt:
		logger.Log.Infoln("app - Run - signal: " + s.String())
	}
}
