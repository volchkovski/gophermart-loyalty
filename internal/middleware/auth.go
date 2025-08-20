package middleware

import (
	"context"
	"github.com/volchkovski/gophermart-loyalty/internal/logger"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
	"net/http"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
)

type TokenVerifier interface {
	VerifyToken(string) (*models.CustomClaims, error)
}

func WithAuth(v TokenVerifier) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		verifyFn := func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")
			claims, err := v.VerifyToken(tokenString)
			if err != nil {
				logger.Log.Infof("Invalid token: %s", err.Error())
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			// Используем наш собственный тип в качестве ключа
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			h.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(verifyFn)
	}
}
