package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/volchkovski/gophermart-loyalty/internal/logger"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
	"github.com/volchkovski/gophermart-loyalty/internal/services/orders"
	"github.com/volchkovski/gophermart-loyalty/internal/valid"
)

type OrderManager interface {
	SaveOrder(ctx context.Context, userID int64, orderNumber string) error
	Orders(ctx context.Context, userID int64) ([]*models.Order, error)
}

func NewOrderHandler(o OrderManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, err := contextUserID(ctx)
		if err != nil {
			logger.Log.Errorln(err.Error())
			handleInternalServerError(w)
			return
		}
		body, err := io.ReadAll(r.Body)
		defer func() {
			if errClose := r.Body.Close(); errClose != nil {
				logger.Log.Errorf("Failed to close request body: %s", errClose.Error())
			}
		}()
		if err != nil {
			logger.Log.Errorln("Failed to read request body: %s", err.Error())
			handleInternalServerError(w)
			return
		}
		orderNumber := string(body)
		if !valid.OrderNumber(orderNumber) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		err = o.SaveOrder(r.Context(), userID, orderNumber)
		if err != nil {
			if errors.Is(err, orders.ErrAlreadyExists) {
				w.WriteHeader(http.StatusOK)
				return
			}
			if errors.Is(err, orders.ErrAnotherUser) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			logger.Log.Errorln(err.Error())
			handleInternalServerError(w)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

func OrdersHandler(o OrderManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, err := contextUserID(ctx)
		if err != nil {
			logger.Log.Errorf("OrdersHandler: Failed to get user ID from context: %s", err.Error())
			handleInternalServerError(w)
			return
		}
		logger.Log.Debugf("OrdersHandler: Getting orders for user %d", userID)
		userOrders, err := o.Orders(ctx, userID)
		if err != nil {
			logger.Log.Errorf("OrdersHandler: Failed to get orders for user %d: %s", userID, err.Error())
			handleInternalServerError(w)
			return
		}
		logger.Log.Debugf("OrdersHandler: Found %d orders for user %d", len(userOrders), userID)
		if len(userOrders) == 0 {
			logger.Log.Debugf("OrdersHandler: No orders found for user %d, returning 204", userID)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(userOrders); err != nil {
			logger.Log.Errorf("OrdersHandler: Failed to encode result for user %d: %s", userID, err.Error())
			handleInternalServerError(w)
			return
		}
		logger.Log.Debugf("OrdersHandler: Successfully returned %d orders for user %d", len(userOrders), userID)
	}
}
