package handlers

import (
	"context"
	"fmt"
)

func contextUserID(ctx context.Context) (int64, error) {
	v := ctx.Value("user_id")
	userID, ok := v.(int64)
	if !ok {
		return 0, fmt.Errorf("user_id %v converting to integer error", v)
	}
	return userID, nil
}
