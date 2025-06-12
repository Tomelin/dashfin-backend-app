// Package mocks provides mock implementations for service interfaces.
package mocks

import (
	"context"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
	"github.com/stretchr/testify/mock"
)

// MockDashboardService is a mock implementation of dashboard.ServiceInterface.
type MockDashboardService struct {
	mock.Mock
}

// GetDashboardSummary mocks the GetDashboardSummary method.
func (m *MockDashboardService) GetDashboardSummary(ctx context.Context, userID string) (*entity_dashboard.DashboardSummary, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity_dashboard.DashboardSummary), args.Error(1)
}
