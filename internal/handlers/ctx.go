package handlers

import (
	"context"
	"fmt"
	"github.com/volchkovski/gophermart-loyalty/internal/middleware"
)

func contextUserID(ctx context.Context) (int64, error) {
	value := ctx.Value(middleware.UserIDKey)
	if value == nil {
		return 0, fmt.Errorf("user_id not found in context")
	}

	userID, ok := value.(int64)
	if !ok {
		return 0, fmt.Errorf("user_id has invalid type, expected int64")
	}

	return userID, nil
}
