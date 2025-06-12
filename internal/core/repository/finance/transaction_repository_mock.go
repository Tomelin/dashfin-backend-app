// Package finance provides mock implementations of repository interfaces.
package finance

import (
	"context"
	"time"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
)

// MockTransactionRepository is a mock implementation of TransactionRepositoryInterface.
type MockTransactionRepository struct{}

// GetMonthlyRevenueByUserID returns fixed monthly revenue for testing.
func (m *MockTransactionRepository) GetMonthlyRevenueByUserID(ctx context.Context, userID string, year int, month int) (float64, error) {
	return 5000.00, nil
}

// GetMonthlyExpensesByUserID returns fixed monthly expenses for testing.
func (m *MockTransactionRepository) GetMonthlyExpensesByUserID(ctx context.Context, userID string, year int, month int) (float64, error) {
	return 2500.00, nil
}

// GetUpcomingBillsByUserID returns a fixed list of upcoming bills for testing.
func (m *MockTransactionRepository) GetUpcomingBillsByUserID(ctx context.Context, userID string, limit int) ([]entity_dashboard.UpcomingBill, error) {
	return []entity_dashboard.UpcomingBill{
		{
			BillID:      "bill_1",
			Description: "Netflix Subscription",
			Amount:      15.99,
			DueDate:     time.Now().AddDate(0, 0, 5).Format("2006-01-02"),
			IsPaid:      false,
			PayNowLink:  "http://example.com/pay/netflix",
		},
		{
			BillID:      "bill_2",
			Description: "Electricity Bill",
			Amount:      75.50,
			DueDate:     time.Now().AddDate(0, 0, 10).Format("2006-01-02"),
			IsPaid:      false,
			PayNowLink:  "http://example.com/pay/electricity",
		},
	}, nil
}

// GetRevenueExpenseChartDataByUserID returns fixed data for the revenue vs. expense chart.
func (m *MockTransactionRepository) GetRevenueExpenseChartDataByUserID(ctx context.Context, userID string, periods int) ([]entity_dashboard.RevenueExpenseChartData, error) {
	data := []entity_dashboard.RevenueExpenseChartData{}
	for i := 0; i < periods; i++ {
		month := time.Now().AddDate(0, -i, 0).Format("Jan")
		data = append(data, entity_dashboard.RevenueExpenseChartData{
			Month:   month,
			Revenue: 6000.00 - float64(i*200), // Decreasing revenue for past months
			Expense: 3000.00 + float64(i*100), // Increasing expense for past months
		})
	}
	return data, nil
}

// GetExpenseCategoryChartDataByUserID returns fixed data for the expense by category chart.
func (m *MockTransactionRepository) GetExpenseCategoryChartDataByUserID(ctx context.Context, userID string, year int, month int) ([]entity_dashboard.ExpenseCategoryChartData, error) {
	return []entity_dashboard.ExpenseCategoryChartData{
		{Category: "Groceries", Amount: 350.75},
		{Category: "Utilities", Amount: 120.00},
		{Category: "Transport", Amount: 85.50},
		{Category: "Entertainment", Amount: 150.25},
		{Category: "Healthcare", Amount: 50.00},
	}, nil
}

// NewMockTransactionRepository creates a new instance of MockTransactionRepository.
func NewMockTransactionRepository() *MockTransactionRepository {
	return &MockTransactionRepository{}
}
