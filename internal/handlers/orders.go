package handlers

import (
	"context"
	"encoding/json"
	"github.com/volchkovski/gophermart-loyalty/internal/logger"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
	"github.com/volchkovski/gophermart-loyalty/internal/utils"
	"io"
	"net/http"
	"strconv"
)

type OrderManager interface {
	SaveOrder(context.Context, *models.Order) (*models.SaveOrderResult, error)
	Orders(context.Context, int) ([]*models.Order, error)
}

func NewOrderHandler(o OrderManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, err := utils.ContextUserID(ctx)
		if err != nil {
			logger.Log.Errorln(err.Error())
			handleInternalServerError(w)
			return
		}
		body, err := io.ReadAll(r.Body)
		defer func() {
			if err := r.Body.Close(); err != nil {
				logger.Log.Errorf("Failed to close request body: %s", err.Error())
			}
		}()
		if err != nil {
			logger.Log.Errorln("Failed to read request body: %s", err.Error())
			handleInternalServerError(w)
			return
		}
		orderNumber, err := strconv.Atoi(string(body))
		if err != nil {
			logger.Log.Errorln("Order number is not string")
			http.Error(w, "Invalid order number", http.StatusBadRequest)
			return
		}
		result, err := o.SaveOrder(r.Context(), &models.Order{Number: orderNumber, UserID: userID})
		if err != nil {
			handleInternalServerError(w)
			return
		}
		if result.Fail != nil {
			http.Error(w, result.Fail.Msg, result.Fail.StatusCode)
			return
		}
		w.WriteHeader(result.StatusCode)
	}
}

func OrdersHandler(o OrderManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, err := utils.ContextUserID(ctx)
		if err != nil {
			logger.Log.Errorln(err.Error())
			handleInternalServerError(w)
			return
		}
		orders, err := o.Orders(ctx, userID)
		if err != nil {
			logger.Log.Errorf("Failed to get orders for user %d: %s", userID, err.Error())
			handleInternalServerError(w)
			return
		}
		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(orders); err != nil {
			logger.Log.Errorf("Failed to encode result for %d user: %s", userID, err.Error())
			handleInternalServerError(w)
			return
		}
	}
}
