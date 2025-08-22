package storage

import "errors"

var (
	ErrUserLoginAlreadyExists = errors.New("user login already exists")
	ErrOrderAlreadyExists     = errors.New("order already exists")
	ErrOrderAnotherUser       = errors.New("order was uploaded by another user")
	ErrLoyaltyNotEnough       = errors.New("loyalty can not be negative")
)
