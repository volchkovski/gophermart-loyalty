package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/volchkovski/gophermart-loyalty/internal/logger"
	"github.com/volchkovski/gophermart-loyalty/internal/services/auth"
)

type Auth interface {
	Register(ctx context.Context, login string, password string) (token string, err error)
	Login(ctx context.Context, login string, password string) (token string, err error)
}

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func RegisterHandler(a Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if errClose := r.Body.Close(); errClose != nil {
				logger.Log.Errorf("Failed to close request body: %s", errClose.Error())
			}
		}()
		c, err := parseCredentials(r)
		if err != nil {
			logger.Log.Errorln("Failed to get credentials from request body: %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		token, err := a.Register(r.Context(), c.Login, c.Password)
		if err != nil {
			if errors.Is(err, auth.ErrLoginIsTaken) {
				logger.Log.Infoln(err.Error())
				w.WriteHeader(http.StatusConflict)
				return
			}
			logger.Log.Errorf("Failed to register user: %s", err.Error())
			handleInternalServerError(w)
			return
		}
		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)
	}
}

func LoginHandler(a Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if errClose := r.Body.Close(); errClose != nil {
				logger.Log.Errorf("Failed to close request body: %s", errClose.Error())
			}
		}()
		c, err := parseCredentials(r)
		if err != nil {
			logger.Log.Errorln("Failed to get credentials from request body: %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		token, err := a.Login(r.Context(), c.Login, c.Password)
		if err != nil {
			if errors.Is(err, auth.ErrNoUser) || errors.Is(err, auth.ErrInvalidPassword) {
				logger.Log.Infoln(err.Error())
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			logger.Log.Errorf("Failed to login: %s", err.Error())
			handleInternalServerError(w)
			return
		}
		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)
	}
}

func parseCredentials(r *http.Request) (*Credentials, error) {
	var c Credentials
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(body, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
