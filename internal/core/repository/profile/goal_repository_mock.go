// Package profile provides mock implementations of repository interfaces.
package profile

import (
	"context"
	"fmt"
	"time"
)

// MockGoalRepository is a mock implementation of GoalRepositoryInterface.
type MockGoalRepository struct{}

// GetGoalsProgressByUserID returns a fixed string representing goals progress for testing.
// As noted in the interface, this might need to be a more complex type in a real scenario.
func (m *MockGoalRepository) GetGoalsProgressByUserID(ctx context.Context, userID string) (string, error) {
	// Example: "Savings Goal: $1500/$5000 (30% complete), Next Milestone: Vacation Fund by Dec 2024"
	// For simplicity, returning a basic string.
	progress := fmt.Sprintf("User %s - Savings Goal for New Car: 60%% complete. Expected by %s.",
		userID, time.Now().AddDate(1, 0, 0).Format("Jan 2006"))
	return progress, nil
}

// NewMockGoalRepository creates a new instance of MockGoalRepository.
func NewMockGoalRepository() *MockGoalRepository {
	return &MockGoalRepository{}
}
