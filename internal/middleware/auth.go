package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/volchkovski/gophermart-loyalty/internal/logger"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
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
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Log.Info("Missing authorization header")
				http.Error(w, "Missing authorization", http.StatusUnauthorized)
				return
			}

			// Извлекаем токен из "Bearer <token>"
			tokenString := authHeader
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}

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
