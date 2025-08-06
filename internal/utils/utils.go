package utils

import (
	"context"
	"fmt"
)

func ContextUserID(ctx context.Context) (int, error) {
	v := ctx.Value("user_id")
	userID, ok := v.(int)
	if !ok {
		return 0, fmt.Errorf("user_id %v converting to integer error", v)
	}
	return userID, nil
}
