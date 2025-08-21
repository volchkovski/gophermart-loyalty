package app

import (
	"context"
	l "log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/volchkovski/gophermart-loyalty/internal/config"
	"github.com/volchkovski/gophermart-loyalty/internal/httpserver"
	"github.com/volchkovski/gophermart-loyalty/internal/logger"
	"github.com/volchkovski/gophermart-loyalty/internal/orderupdater"
	"github.com/volchkovski/gophermart-loyalty/internal/router"
	"github.com/volchkovski/gophermart-loyalty/internal/services/auth"
	"github.com/volchkovski/gophermart-loyalty/internal/services/loyalty"
	"github.com/volchkovski/gophermart-loyalty/internal/services/orders"
	"github.com/volchkovski/gophermart-loyalty/internal/storage/pg"
)

func MustRun(cfg *config.Config) {
	if err := logger.Initialize(cfg.LogLevel, cfg.Env); err != nil {
		panic(err)
	}

	defer func() {
		if errSync := logger.Log.Sync(); errSync != nil {
			l.Printf("Failed flush logger buffer: %s", errSync.Error())
		}
	}()

	db, err := pg.New(cfg.DSN)
	if err != nil {
		panic(err.Error())
	}
	defer func() {
		if errClose := db.Close(); errClose != nil {
			logger.Log.Warnf("Failed to close db conn: %s", errClose.Error())
		}
	}()
	type (
		Auth           = auth.Auth
		LoyaltyManager = loyalty.Manager
		OrdersManager  = orders.Manager
	)
	p := struct {
		*Auth
		*LoyaltyManager
		*OrdersManager
	}{
		Auth:           auth.New(cfg.Secret, db),
		LoyaltyManager: loyalty.NewManager(db),
		OrdersManager:  orders.NewManager(db),
	}

	r := router.NewHTTPRouter(&p)
	server := httpserver.New(cfg.Addr, r)
	server.Start()

	updater := orderupdater.New(db, cfg.AccrualAddr)
	updaterCtx, updaterCancel := context.WithCancel(context.Background())
	defer updaterCancel()
	updater.Start(updaterCtx)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case err = <-server.Notify():
		logger.Log.Errorf("HTTP server error: %s", err.Error())
	case err = <-updater.Notify():
		logger.Log.Errorf("Order updater error: %s", err.Error())
	case s := <-interrupt:
		logger.Log.Infof("Received signal: %s, starting graceful shutdown", s.String())
		gracefulShutdown(server, updaterCancel, 30*time.Second)
		return
	}
	if err != nil {
		logger.Log.Errorf("Application error, performing graceful shutdown: %s", err.Error())
		gracefulShutdown(server, updaterCancel, 10*time.Second)
	}
}

func gracefulShutdown(s *httpserver.HTTPServer, updaterCancel context.CancelFunc, serverTimeOut time.Duration) {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), serverTimeOut)
	defer shutdownCancel()

	updaterCancel()
	logger.Log.Info("Order updater shutdown signal sent")

	if shutdownErr := s.Shutdown(shutdownCtx); shutdownErr != nil {
		logger.Log.Errorf("HTTP server shutdown failed: %s", shutdownErr.Error())
	}

	logger.Log.Info("Application shutdown completed")
}
