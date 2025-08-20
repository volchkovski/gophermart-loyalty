package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
	"github.com/volchkovski/gophermart-loyalty/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type db interface {
	User(ctx context.Context, login string) (*models.User, error)
	NewUser(ctx context.Context, login, passwordHash string) (*models.User, error)
}

type Auth struct {
	secret []byte
	db     db
}

func New(secret string, db db) *Auth {
	return &Auth{
		secret: []byte(secret),
		db:     db,
	}
}

func (a *Auth) Register(ctx context.Context, login string, password string) (string, error) {
	//user, err := a.db.User(ctx, login)
	//if err != nil {
	//	return "", fmt.Errorf("user_id check error: %w", err)
	//}
	//if user != nil {
	//	return "", ErrLoginIsTaken
	//}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("password hash gen error: %w", err)
	}
	user, err := a.db.NewUser(ctx, login, string(passwordHash))
	if err != nil {
		if errors.Is(err, storage.ErrUserLoginAlreadyExists) {
			return "", ErrLoginIsTaken
		}
		return "", fmt.Errorf("new user creation error: %w", err)
	}
	token, err := a.createToken(user.ID)
	if err != nil {
		return "", fmt.Errorf("token creation error: %w", err)
	}
	return token, nil
}

func (a *Auth) Login(ctx context.Context, login string, password string) (string, error) {
	user, err := a.db.User(ctx, login)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return "", ErrNoUser
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		return "", ErrInvalidPassword
	}
	token, err := a.createToken(user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
	}
	return token, nil
}

func (a *Auth) createToken(userID int64) (string, error) {
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, models.CustomClaims{
		UserID: userID,
	}).SignedString(a.secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (a *Auth) VerifyToken(tokenString string) (*models.CustomClaims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("empty token")
	}
	token, err := jwt.ParseWithClaims(tokenString, &models.CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return a.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*models.CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
