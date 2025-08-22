package orders

import "errors"

var (
	ErrAlreadyExists = errors.New("order already exists")
	ErrAnotherUser   = errors.New("order was uploaded by another user")
)
