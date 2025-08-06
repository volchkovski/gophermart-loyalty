package auth

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type authDB interface {
	User(context.Context, string) (*models.User, error)
	NewUser(context.Context, string, string) (*models.User, error)
}

type Auth struct {
	secret []byte
	db     authDB
}

func New(secret string, db authDB) *Auth {
	return &Auth{
		secret: []byte(secret),
		db:     db,
	}
}

func (a *Auth) Register(ctx context.Context, data *models.RegistrationData) (*models.RegistrationResult, error) {
	user, err := a.db.User(ctx, data.Login)
	if err != nil {
		return nil, fmt.Errorf("user_id check error: %w", err)
	}
	if user != nil {
		return &models.RegistrationResult{
			Token: "",
			Fail: &models.Fail{
				Msg:        "login already exists",
				StatusCode: http.StatusConflict,
			},
		}, nil
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("password hash gen error: %w", err)
	}
	user, err = a.db.NewUser(ctx, data.Login, string(passwordHash))
	if err != nil {
		return nil, fmt.Errorf("new user creation error: %w", err)
	}
	token, err := a.createToken(user)
	if err != nil {
		return nil, fmt.Errorf("token creation error: %w", err)
	}
	return &models.RegistrationResult{
		Token: token,
		Fail:  nil,
	}, nil
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

func (a *Auth) Login(ctx context.Context, d *models.RegistrationData) (*models.LoggingResult, error) {
	user, err := a.db.User(ctx, d.Login)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return &models.LoggingResult{
			Token: "",
			Fail: &models.Fail{
				Msg:        "login is not registered",
				StatusCode: http.StatusUnauthorized,
			},
		}, nil
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(d.Password)); err != nil {
		return &models.LoggingResult{
			Token: "",
			Fail: &models.Fail{
				Msg:        "invalid login or password",
				StatusCode: http.StatusUnauthorized,
			},
		}, nil
	}
	token, err := a.createToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}
	return &models.LoggingResult{
		Token: token,
		Fail:  nil,
	}, nil
}

func (a *Auth) createToken(u *models.User) (string, error) {
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, models.CustomClaims{
		UserID: u.ID,
	}).SignedString(a.secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
