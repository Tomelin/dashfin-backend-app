// Package finance provides mock implementations of service interfaces.
package finance

import (
	"context"
	"time"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
)

// MockTransactionService is a mock implementation of TransactionServiceInterface.
type MockTransactionService struct{}

// GetMonthlyRevenueByUserID returns fixed monthly revenue for testing.
func (m *MockTransactionService) GetMonthlyRevenueByUserID(ctx context.Context, userID string, year int, month int) (float64, error) {
	return 5500.00, nil // Slightly different from repo mock for distinction
}

// GetMonthlyExpensesByUserID returns fixed monthly expenses for testing.
func (m *MockTransactionService) GetMonthlyExpensesByUserID(ctx context.Context, userID string, year int, month int) (float64, error) {
	return 2200.00, nil // Slightly different
}

// GetUpcomingBillsByUserID returns a fixed list of upcoming bills for testing.
func (m *MockTransactionService) GetUpcomingBillsByUserID(ctx context.Context, userID string, limit int) ([]entity_dashboard.UpcomingBill, error) {
	return []entity_dashboard.UpcomingBill{
		{
			BillID:      "serv_bill_1",
			Description: "Service Internet Bill",
			Amount:      60.00,
			DueDate:     time.Now().AddDate(0, 0, 7).Format("2006-01-02"),
			IsPaid:      false,
			PayNowLink:  "http://example.com/pay/service_internet",
		},
	}, nil
}

// GetRevenueExpenseChartDataByUserID returns fixed data for the revenue vs. expense chart.
func (m *MockTransactionService) GetRevenueExpenseChartDataByUserID(ctx context.Context, userID string, periods int) ([]entity_dashboard.RevenueExpenseChartData, error) {
	data := []entity_dashboard.RevenueExpenseChartData{}
	for i := 0; i < periods; i++ {
		month := time.Now().AddDate(0, -i, 0).Format("Jan") // Corrected to use "Jan" format
		data = append(data, entity_dashboard.RevenueExpenseChartData{
			Month:   month,
			Revenue: 6500.00 - float64(i*250),
			Expense: 2800.00 + float64(i*150),
		})
	}
	return data, nil
}

// GetExpenseCategoryChartDataByUserID returns fixed data for the expense by category chart.
func (m *MockTransactionService) GetExpenseCategoryChartDataByUserID(ctx context.Context, userID string, year int, month int) ([]entity_dashboard.ExpenseCategoryChartData, error) {
	return []entity_dashboard.ExpenseCategoryChartData{
		{Category: "Service Groceries", Amount: 320.50},
		{Category: "Service Utilities", Amount: 110.00},
		{Category: "Service Transport", Amount: 75.25},
	}, nil
}

// NewMockTransactionService creates a new instance of MockTransactionService.
func NewMockTransactionService() *MockTransactionService {
	return &MockTransactionService{}
}
