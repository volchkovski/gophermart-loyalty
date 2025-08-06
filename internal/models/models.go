package models

import "github.com/golang-jwt/jwt/v5"

type Fail struct {
	Msg        string
	StatusCode int
}

type RegistrationData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegistrationResult struct {
	Token string
	Fail  *Fail
}

type LoggingResult struct {
	Token string
	Fail  *Fail
}

type Order struct {
	Number     int    `json:"number"`
	Status     string `json:"status"`
	Accrual    int    `json:"accrual"`
	UploadedAt int    `json:"-"`
	UserID     int    `json:"-"`
}

type SaveOrderResult struct {
	StatusCode int
	Fail       *Fail
}

type BalanceResult struct {
	Current   int `json:"-"`
	Withdrawn int `json:"withdrawn"`
}

type Withdrawal struct {
	Order       int `json:"-"`
	Sum         int `json:"sum"`
	ProcessedAt int `json:"-"`
}

type WithdrawResult struct {
	Fail *Fail
}

type User struct {
	ID             int
	HashedPassword string
}

type CustomClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}
