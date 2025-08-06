package handlers

import (
	"context"
	"encoding/json"
	"github.com/volchkovski/gophermart-loyalty/internal/logger"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
	"io"
	"net/http"
)

type BalanceManager interface {
	Balance(context.Context, int) (models.BalanceResult, error)
	Withdraw(context.Context, models.Withdrawal) (models.WithdrawResult, error)
	Withdrawals(context.Context, int) ([]*models.Withdrawal, error)
}

func BalanceHandler(b BalanceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, ok := ctx.Value("user_id").(int)
		if !ok {
			logger.Log.Errorln("Failed to convert user_id to integer from context")
			handleInternalServerError(w)
			return
		}
		result, err := b.Balance(ctx, userID)
		if err != nil {
			logger.Log.Errorf("Failed to get balance for user %d: %s", userID, err.Error())
			handleInternalServerError(w)
			return
		}
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(result); err != nil {
			logger.Log.Errorf("Failed to encode result for %d user: %s", userID, err.Error())
			handleInternalServerError(w)
			return
		}
	}
}

func WithdrawHandler(b BalanceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var withdrawal models.Withdrawal
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
		if err = json.Unmarshal(body, &withdrawal); err != nil {
			logger.Log.Errorln("Failed to unmarshal request body: %s", err.Error())
			handleInternalServerError(w)
			return
		}
		result, err := b.Withdraw(r.Context(), withdrawal)
		if err != nil {
			logger.Log.Errorf("Failed to withdraw: %s", err.Error())
			return
		}
		if result.Fail != nil {
			logger.Log.Infoln("Unsuccessful withdrawal")
			http.Error(w, result.Fail.Msg, result.Fail.StatusCode)
			return
		}
	}
}

func WithdrawalsHandler(b BalanceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, ok := ctx.Value("user_id").(int)
		if !ok {
			logger.Log.Errorln("Failed to convert user_id to integer from context")
			handleInternalServerError(w)
			return
		}
		withdrawals, err := b.Withdrawals(r.Context(), userID)
		if err != nil {
			logger.Log.Errorf("Failed to withdrawals for user %d: %s", userID, err.Error())
			handleInternalServerError(w)
			return
		}
		if len(withdrawals) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if err = json.NewEncoder(w).Encode(&withdrawals); err != nil {
			logger.Log.Errorf("Failed to encode withdrawals for user with id %d: %s", userID, err.Error())
			handleInternalServerError(w)
			return
		}
	}
}
