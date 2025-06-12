// Package profile defines repository interfaces for profile-related data.
package profile

import (
	"context"
)

// GoalRepositoryInterface defines methods for interacting with goal data.
type GoalRepositoryInterface interface {
	// GetGoalsProgressByUserID retrieves the progress of goals for a given user.
	// Note: The return type is string as specified, but might need adjustment
	// based on how "goals progress" is structured.
	GetGoalsProgressByUserID(ctx context.Context, userID string) (string, error)
}
