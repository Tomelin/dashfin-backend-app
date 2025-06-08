package repository

import (
	"context"
	"errors"
	"fmt"
)

func SetCollection(ctx context.Context, collection string) (*string, error) {

	var col string
	if ctx.Value("UserID") == nil {
		return nil, errors.New("user id is nil")
	} else {
		userID := ctx.Value("UserID").(string)
		if userID == "" {
			return nil, errors.New("user id is empty")
		}
		col = fmt.Sprintf("%s/%s", userID, collection)
	}

	return &col, nil
}
