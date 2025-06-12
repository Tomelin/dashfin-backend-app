// Package profile defines service interfaces for profile-related business logic.
package profile

import (
	"context"
)

// GoalServiceInterface defines methods for goal-related business logic.
type GoalServiceInterface interface {
	// GetGoalsProgressByUserID retrieves the progress of goals for a given user.
	// Note: The return type is string as specified, but might need adjustment
	// based on how "goals progress" is structured.
	GetGoalsProgressByUserID(ctx context.Context, userID string) (string, error)
}
