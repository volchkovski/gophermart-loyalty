package models

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    int64     `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type UnprocessedOrder struct {
	Number string
	UserID int64
}

type Balance struct {
	Current   int64 `json:"-"`
	Withdrawn int64 `json:"-"`
}

type Withdrawal struct {
	Order       string    `json:"order"`
	Sum         int64     `json:"-"`
	ProcessedAt time.Time `json:"processed_at"`
}

type User struct {
	ID             int64
	Login          string
	HashedPassword string
}

type CustomClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}
