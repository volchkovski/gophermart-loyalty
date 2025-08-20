package auth

import "errors"

var (
	ErrLoginIsTaken    = errors.New("login is already taken")
	ErrNoUser          = errors.New("no user with passed login")
	ErrInvalidPassword = errors.New("password mismatch with saved hash")
)
