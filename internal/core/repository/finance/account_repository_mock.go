// Package finance provides mock implementations of repository interfaces.
package finance

import (
	"context"
	"time"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
)

// MockAccountRepository is a mock implementation of AccountRepositoryInterface.
type MockAccountRepository struct{}

// GetTotalBalanceByUserID returns a fixed total balance for testing.
func (m *MockAccountRepository) GetTotalBalanceByUserID(ctx context.Context, userID string) (float64, error) {
	return 12345.67, nil
}

// GetAccountSummariesByUserID returns a fixed list of account summaries for testing.
func (m *MockAccountRepository) GetAccountSummariesByUserID(ctx context.Context, userID string) ([]entity_dashboard.AccountSummaryData, error) {
	return []entity_dashboard.AccountSummaryData{
		{
			AccountID:   "acc_1",
			AccountName: "Savings Account",
			Balance:     10000.00,
			BankName:    "Mock Bank",
			AccountType: "savings",
		},
		{
			AccountID:   "acc_2",
			AccountName: "Checking Account",
			Balance:     2345.67,
			BankName:    "Mock Bank",
			AccountType: "checking",
		},
		{
			AccountID:   "acc_3",
			AccountName: "Credit Card",
			Balance:     -500.50,
			BankName:    "Mock Credit",
			AccountType: "credit card",
		},
	}, nil
}

// NewMockAccountRepository creates a new instance of MockAccountRepository.
func NewMockAccountRepository() *MockAccountRepository {
	return &MockAccountRepository{}
}
