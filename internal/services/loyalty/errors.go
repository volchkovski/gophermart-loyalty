package loyalty

import "errors"

var (
	ErrNotEnough = errors.New("not enough loyalty points to withdraw")
)
