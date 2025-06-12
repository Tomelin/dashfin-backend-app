// Package finance defines service interfaces for finance-related business logic.
package finance

import (
	"context"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
)

// TransactionServiceInterface defines methods for transaction-related business logic.
type TransactionServiceInterface interface {
	// GetMonthlyRevenueByUserID retrieves the total revenue for a given user and month.
	GetMonthlyRevenueByUserID(ctx context.Context, userID string, year int, month int) (float64, error)
	// GetMonthlyExpensesByUserID retrieves the total expenses for a given user and month.
	GetMonthlyExpensesByUserID(ctx context.Context, userID string, year int, month int) (float64, error)
	// GetUpcomingBillsByUserID retrieves a list of upcoming bills for a given user.
	GetUpcomingBillsByUserID(ctx context.Context, userID string, limit int) ([]entity_dashboard.UpcomingBill, error)
	// GetRevenueExpenseChartDataByUserID retrieves data for the revenue vs. expense chart.
	GetRevenueExpenseChartDataByUserID(ctx context.Context, userID string, periods int) ([]entity_dashboard.RevenueExpenseChartData, error)
	// GetExpenseCategoryChartDataByUserID retrieves data for the expense by category chart.
	GetExpenseCategoryChartDataByUserID(ctx context.Context, userID string, year int, month int) ([]entity_dashboard.ExpenseCategoryChartData, error)
}
