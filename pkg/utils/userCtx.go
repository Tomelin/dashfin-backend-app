package utils

import (
	"context"
	"fmt"
)

func GetUserID(ctx context.Context) (*string, error) {
	userIDFromCtx := ctx.Value("UserID")
	if userIDFromCtx == nil {
		return nil, fmt.Errorf("userID not found in context")
	}
	userID, ok := userIDFromCtx.(string)
	if !ok {
		return nil, fmt.Errorf("userID in context is not a string")
	}
	if userID == "" {
		return nil, fmt.Errorf("userID in context is empty")
	}

	return &userID, nil
}
