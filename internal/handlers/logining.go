package handlers

import (
	"context"
	"encoding/json"
	"github.com/volchkovski/gophermart-loyalty/internal/logger"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
	"io"
	"net/http"
)

type Loginer interface {
	Login(context.Context, *models.RegistrationData) (*models.LoggingResult, error)
}

func LoginHandler(l Loginer) http.HandlerFunc {
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
		result, err := l.Login(r.Context(), d)
		if err != nil {
			logger.Log.Errorf("Failed to login: %s", err.Error())
			handleInternalServerError(w)
			return
		}
		if result.Fail != nil {
			logger.Log.Infof("Logining fail for user %s: %s", d.Login, result.Fail.Msg)
			return
		}
		w.Header().Set("Authorization", result.Token)
		w.WriteHeader(http.StatusOK)
	}
}
