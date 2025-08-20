package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/volchkovski/gophermart-loyalty/internal/handlers"
	mw "github.com/volchkovski/gophermart-loyalty/internal/middleware"
)

type HTTPRouter struct {
	chi.Router
}

type Processor interface {
	handlers.Auth
	handlers.OrderManager
	handlers.LoyaltyManager
	mw.TokenVerifier
}

// NewHTTPRouter создает новый маршрутизатор с заданным процессором
func NewHTTPRouter(p Processor) *HTTPRouter {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", handlers.RegisterHandler(p))
		r.Post("/api/user/login", handlers.LoginHandler(p))
	})

	r.Group(func(r chi.Router) {
		r.Use(mw.WithAuth(p))

		r.Route("/api/user/orders", func(r chi.Router) {
			r.Get("/", handlers.OrdersHandler(p))
			r.Post("/", handlers.NewOrderHandler(p))
		})

		r.Route("/api/user/loyalty", func(r chi.Router) {
			r.Get("/", handlers.BalanceHandler(p))
			r.Post("/withdraw", handlers.WithdrawHandler(p))
		})

		r.Get("/api/user/withdrawals", handlers.WithdrawalsHandler(p))
	})

	return &HTTPRouter{Router: r}
}
