// Package finance defines service interfaces for finance-related business logic.
package finance

import (
	"context"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
)

// AccountServiceInterface defines methods for account-related business logic.
type AccountServiceInterface interface {
	// GetTotalBalanceByUserID retrieves the total balance for a given user.
	GetTotalBalanceByUserID(ctx context.Context, userID string) (float64, error)
	// GetAccountSummariesByUserID retrieves a list of account summaries for a given user.
	GetAccountSummariesByUserID(ctx context.Context, userID string) ([]entity_dashboard.AccountSummaryData, error)
}
