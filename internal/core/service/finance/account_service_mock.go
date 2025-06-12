// Package finance provides mock implementations of service interfaces.
package finance

import (
	"context"
	"time"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
)

// MockAccountService is a mock implementation of AccountServiceInterface.
type MockAccountService struct{}

// GetTotalBalanceByUserID returns a fixed total balance for testing.
func (m *MockAccountService) GetTotalBalanceByUserID(ctx context.Context, userID string) (float64, error) {
	// In a real service, this might call a repository. Here, it returns mock data.
	return 12345.67, nil
}

// GetAccountSummariesByUserID returns a fixed list of account summaries for testing.
func (m *MockAccountService) GetAccountSummariesByUserID(ctx context.Context, userID string) ([]entity_dashboard.AccountSummaryData, error) {
	// In a real service, this might call a repository and apply business logic.
	return []entity_dashboard.AccountSummaryData{
		{
			AccountID:   "serv_acc_1",
			AccountName: "Service Savings Account",
			Balance:     10000.00,
			BankName:    "Mock Service Bank",
			AccountType: "savings",
		},
		{
			AccountID:   "serv_acc_2",
			AccountName: "Service Checking Account",
			Balance:     2345.67,
			BankName:    "Mock Service Bank",
			AccountType: "checking",
		},
	}, nil
}

// NewMockAccountService creates a new instance of MockAccountService.
func NewMockAccountService() *MockAccountService {
	return &MockAccountService{}
}
