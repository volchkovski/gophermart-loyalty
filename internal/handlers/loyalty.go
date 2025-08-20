package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/volchkovski/gophermart-loyalty/internal/logger"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
	"github.com/volchkovski/gophermart-loyalty/internal/services/loyalty"
	"github.com/volchkovski/gophermart-loyalty/internal/valid"
)

type LoyaltyManager interface {
	Balance(ctx context.Context, userID int64) (*models.Balance, error)
	Withdraw(ctx context.Context, userID, sum int64, orderNumber string) error
	Withdrawals(context.Context, int64) ([]*models.Withdrawal, error)
}

func BalanceHandler(lm LoyaltyManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, err := contextUserID(ctx)
		if err != nil {
			logger.Log.Errorln(err.Error())
			handleInternalServerError(w)
			return
		}
		b, err := lm.Balance(ctx, userID)
		if err != nil {
			logger.Log.Errorf("Failed to get loyalty for user %d: %s", userID, err.Error())
			handleInternalServerError(w)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(b); err != nil {
			logger.Log.Errorf("Failed to encode result for %d user: %s", userID, err.Error())
			handleInternalServerError(w)
			return
		}
	}
}

func WithdrawHandler(lm LoyaltyManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, err := contextUserID(ctx)
		if err != nil {
			logger.Log.Errorln(err.Error())
			handleInternalServerError(w)
			return
		}
		var wd models.Withdrawal
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
		if err = json.Unmarshal(body, &wd); err != nil {
			logger.Log.Errorln("Failed to unmarshal request body: %s", err.Error())
			handleInternalServerError(w)
			return
		}
		if !valid.OrderNumber(wd.Order) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if err = lm.Withdraw(r.Context(), userID, wd.Sum, wd.Order); err != nil {
			if errors.Is(err, loyalty.ErrNotEnough) {
				w.WriteHeader(http.StatusPaymentRequired)
				return
			}
			logger.Log.Errorf("Failed to withdraw: %s", err.Error())
			handleInternalServerError(w)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func WithdrawalsHandler(lm LoyaltyManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, err := contextUserID(ctx)
		if err != nil {
			logger.Log.Errorln(err.Error())
			handleInternalServerError(w)
			return
		}
		withdrawals, err := lm.Withdrawals(r.Context(), userID)
		if err != nil {
			logger.Log.Errorf("Failed to withdrawals for user %d: %s", userID, err.Error())
			handleInternalServerError(w)
			return
		}
		if len(withdrawals) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(&withdrawals); err != nil {
			logger.Log.Errorf("Failed to encode withdrawals for user with id %d: %s", userID, err.Error())
			handleInternalServerError(w)
			return
		}
	}
}
