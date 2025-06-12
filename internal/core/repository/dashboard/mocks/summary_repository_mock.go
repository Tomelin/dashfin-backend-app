// Package mocks provides mock implementations for repository interfaces.
package mocks

import (
	"context"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
	"github.com/stretchr/testify/mock"
)

// MockDashboardRepository is a mock implementation of dashboard.RepositoryInterface.
type MockDashboardRepository struct {
	mock.Mock
}

// GetSummaryCardsData mocks the GetSummaryCardsData method.
func (m *MockDashboardRepository) GetSummaryCardsData(ctx context.Context, userID string) (*entity_dashboard.SummaryCards, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity_dashboard.SummaryCards), args.Error(1)
}

// GetAccountSummaries mocks the GetAccountSummaries method.
func (m *MockDashboardRepository) GetAccountSummaries(ctx context.Context, userID string) ([]entity_dashboard.AccountSummaryData, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity_dashboard.AccountSummaryData), args.Error(1)
}

// GetUpcomingBills mocks the GetUpcomingBills method.
func (m *MockDashboardRepository) GetUpcomingBills(ctx context.Context, userID string) ([]entity_dashboard.UpcomingBill, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity_dashboard.UpcomingBill), args.Error(1)
}

// GetRevenueExpenseChartData mocks the GetRevenueExpenseChartData method.
func (m *MockDashboardRepository) GetRevenueExpenseChartData(ctx context.Context, userID string) ([]entity_dashboard.RevenueExpenseChartData, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity_dashboard.RevenueExpenseChartData), args.Error(1)
}

// GetExpenseCategoryChartData mocks the GetExpenseCategoryChartData method.
func (m *MockDashboardRepository) GetExpenseCategoryChartData(ctx context.Context, userID string) ([]entity_dashboard.ExpenseCategoryChartData, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity_dashboard.ExpenseCategoryChartData), args.Error(1)
}

// GetPersonalizedRecommendations mocks the GetPersonalizedRecommendations method.
func (m *MockDashboardRepository) GetPersonalizedRecommendations(ctx context.Context, userID string) ([]entity_dashboard.PersonalizedRecommendation, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity_dashboard.PersonalizedRecommendation), args.Error(1)
}
