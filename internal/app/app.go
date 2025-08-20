package app

import (
	"context"
	"github.com/volchkovski/gophermart-loyalty/internal/config"
	"github.com/volchkovski/gophermart-loyalty/internal/httpserver"
	"github.com/volchkovski/gophermart-loyalty/internal/logger"
	"github.com/volchkovski/gophermart-loyalty/internal/orderupdater"
	"github.com/volchkovski/gophermart-loyalty/internal/router"
	"github.com/volchkovski/gophermart-loyalty/internal/services/auth"
	"github.com/volchkovski/gophermart-loyalty/internal/services/loyalty"
	"github.com/volchkovski/gophermart-loyalty/internal/services/orders"
	"github.com/volchkovski/gophermart-loyalty/internal/storage/pg"
	l "log"
	"os"
	"os/signal"
	"syscall"
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
			logger.Log.Warnf("Failed to close db conn: %s", err.Error())
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	updater.Start(ctx)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case err = <-server.Notify():
	case err = <-updater.Notify():
	case s := <-interrupt:
		logger.Log.Infof("app - MustRun - signal: %s", s.String())
	}
	if err != nil {
		logger.Log.Infof("app - MustRun - error: %s", err.Error())
	}
}
