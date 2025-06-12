// Package profile provides mock implementations of service interfaces.
package profile

import (
	"context"
	"fmt"
	"time"
)

// MockGoalService is a mock implementation of GoalServiceInterface.
type MockGoalService struct{}

// GetGoalsProgressByUserID returns a fixed string representing goals progress for testing.
// This mock service might add a bit more detail or formatting than the repository mock.
func (m *MockGoalService) GetGoalsProgressByUserID(ctx context.Context, userID string) (string, error) {
	// Example: "Service: User JohnDoe - Current Goal: Emergency Fund ($2500/$10000) - 25% Achieved. Keep going!"
	// For simplicity, returning a slightly different string than the repository mock.
	progress := fmt.Sprintf("Service Update for User %s: Your 'Rainy Day Fund' is 75%% complete. Target date: %s.",
		userID, time.Now().AddDate(0, 6, 0).Format("02 Jan 2006"))
	return progress, nil
}

// NewMockGoalService creates a new instance of MockGoalService.
func NewMockGoalService() *MockGoalService {
	return &MockGoalService{}
}
