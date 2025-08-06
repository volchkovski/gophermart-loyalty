package handlers

import (
	"context"
	"encoding/json"
	"github.com/volchkovski/gophermart-loyalty/internal/logger"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
	"io"
	"net/http"
)

type Registerer interface {
	Register(context.Context, *models.RegistrationData) (*models.RegistrationResult, error)
}

func RegisterHandler(reg Registerer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		d := new(models.RegistrationData)
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
		if err = json.Unmarshal(body, d); err != nil {
			logger.Log.Errorln("Failed to unmarshal request body: %s", err.Error())
			handleInternalServerError(w)
			return
		}
		result, err := reg.Register(r.Context(), d)
		if err != nil {
			logger.Log.Errorf("Failed to register user: %s", err.Error())
			handleInternalServerError(w)
			return
		}
		if result.Fail != nil {
			logger.Log.Infof("Registration fail for user %s: %s", d.Login, result.Fail.Msg)
			http.Error(w, result.Fail.Msg, result.Fail.StatusCode)
			return
		}
		w.Header().Set("Authorization", result.Token)
		w.WriteHeader(http.StatusOK)
	}
}
