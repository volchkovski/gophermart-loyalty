package middleware

import (
	"context"
	"github.com/volchkovski/gophermart-loyalty/internal/logger"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
	"net/http"
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
				handleInvalidToken(w)
				return
			}
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			h.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(verifyFn)
	}
}

func handleInvalidToken(w http.ResponseWriter) {
	http.Error(w, "Invalid token", http.StatusUnauthorized)
}
