package finance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	financeEntity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	profileEntity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
)

// MockBankAccountService is a mock type for BankAccountServiceInterface
type MockBankAccountService struct {
	mock.Mock
}

func (m *MockBankAccountService) CreateBankAccount(ctx context.Context, data *financeEntity.BankAccount) (*financeEntity.BankAccountRequest, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*financeEntity.BankAccountRequest), args.Error(1)
}

func (m *MockBankAccountService) GetBankAccountByID(ctx context.Context, id *string) (*financeEntity.BankAccountRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*financeEntity.BankAccountRequest), args.Error(1)
}

func (m *MockBankAccountService) GetBankAccounts(ctx context.Context) ([]financeEntity.BankAccountRequest, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]financeEntity.BankAccountRequest), args.Error(1)
}
func (m *MockBankAccountService) GetByFilter(ctx context.Context, data map[string]interface{}) ([]financeEntity.BankAccountRequest, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]financeEntity.BankAccountRequest), args.Error(1)
}
func (m *MockBankAccountService) UpdateBankAccount(ctx context.Context, data *financeEntity.BankAccountRequest) (*financeEntity.BankAccountRequest, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*financeEntity.BankAccountRequest), args.Error(1)
}
func (m *MockBankAccountService) DeleteBankAccount(ctx context.Context, id *string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}


// MockExpenseRecordService is a mock type for ExpenseRecordServiceInterface
type MockExpenseRecordService struct {
	mock.Mock
}

func (m *MockExpenseRecordService) CreateExpenseRecord(ctx context.Context, data *financeEntity.ExpenseRecord) (*financeEntity.ExpenseRecord, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*financeEntity.ExpenseRecord), args.Error(1)
}

func (m *MockExpenseRecordService) GetExpenseRecordByID(ctx context.Context, id string) (*financeEntity.ExpenseRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*financeEntity.ExpenseRecord), args.Error(1)
}

func (m *MockExpenseRecordService) GetExpenseRecords(ctx context.Context) ([]financeEntity.ExpenseRecord, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]financeEntity.ExpenseRecord), args.Error(1)
}

func (m *MockExpenseRecordService) GetExpenseRecordsByFilter(ctx context.Context, filter map[string]interface{}) ([]financeEntity.ExpenseRecord, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]financeEntity.ExpenseRecord), args.Error(1)
}

func (m *MockExpenseRecordService) GetExpenseRecordsByDate(ctx context.Context, filter *financeEntity.ExpenseRecordQueryByDate) ([]financeEntity.ExpenseRecord, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]financeEntity.ExpenseRecord), args.Error(1)
}

func (m *MockExpenseRecordService) UpdateExpenseRecord(ctx context.Context, id string, data *financeEntity.ExpenseRecord) (*financeEntity.ExpenseRecord, error) {
	args := m.Called(ctx, id, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*financeEntity.ExpenseRecord), args.Error(1)
}

func (m *MockExpenseRecordService) DeleteExpenseRecord(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockIncomeRecordService is a mock type for IncomeRecordServiceInterface
type MockIncomeRecordService struct {
	mock.Mock
}

func (m *MockIncomeRecordService) CreateIncomeRecord(ctx context.Context, data *financeEntity.IncomeRecord) (*financeEntity.IncomeRecord, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*financeEntity.IncomeRecord), args.Error(1)
}
func (m *MockIncomeRecordService) GetIncomeRecordByID(ctx context.Context, id string) (*financeEntity.IncomeRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*financeEntity.IncomeRecord), args.Error(1)
}
func (m *MockIncomeRecordService) GetIncomeRecords(ctx context.Context, params *financeEntity.GetIncomeRecordsQueryParameters) ([]financeEntity.IncomeRecord, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]financeEntity.IncomeRecord), args.Error(1)
}
func (m *MockIncomeRecordService) UpdateIncomeRecord(ctx context.Context, id string, data *financeEntity.IncomeRecord) (*financeEntity.IncomeRecord, error) {
	args := m.Called(ctx, id, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*financeEntity.IncomeRecord), args.Error(1)
}
func (m *MockIncomeRecordService) DeleteIncomeRecord(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}


// MockProfileGoalsService is a mock type for ProfileGoalsServiceInterface
type MockProfileGoalsService struct {
	mock.Mock
}

func (m *MockProfileGoalsService) UpdateProfileGoals(ctx context.Context, userId *string, data *profileEntity.ProfileGoals) (*profileEntity.ProfileGoals, error) {
	args := m.Called(ctx, userId, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*profileEntity.ProfileGoals), args.Error(1)
}

func (m *MockProfileGoalsService) GetProfileGoals(ctx context.Context, userID *string) (profileEntity.ProfileGoals, error) {
	// Note: ProfileGoals is a struct, not a pointer, in the interface.
	args := m.Called(ctx, userID)
	return args.Get(0).(profileEntity.ProfileGoals), args.Error(1)
}


func TestDashboardService_GetDashboardData_Nominal(t *testing.T) {
	mockBankSvc := new(MockBankAccountService)
	mockExpenseSvc := new(MockExpenseRecordService)
	mockIncomeSvc := new(MockIncomeRecordService)
	mockGoalsSvc := new(MockProfileGoalsService)

	dashboardService := NewDashboardService(mockBankSvc, mockExpenseSvc, mockIncomeSvc, mockGoalsSvc)

	ctx := context.WithValue(context.Background(), "UserID", "test-user-123")
	now := time.Now()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// --- Mocking BankAccountService ---
	mockAccounts := []financeEntity.BankAccountRequest{
		{ID: "acc1", BankAccount: financeEntity.BankAccount{CustomBankName: "Conta Corrente A", AccountNumber: "111"}},
		{ID: "acc2", BankAccount: financeEntity.BankAccount{CustomBankName: "Poupança B", AccountNumber: "222"}},
	}
	mockBankSvc.On("GetBankAccounts", ctx).Return(mockAccounts, nil)

	// --- Mocking IncomeRecordService ---
	// Income for acc1 (current month)
	// Income for acc2 (previous month)
	mockIncomes := []financeEntity.IncomeRecord{
		{BankAccountID: "acc1", Amount: 2000.00, ReceiptDate: currentMonthStart.Format("2006-01-02")},
		{BankAccountID: "acc1", Amount: 3000.00, ReceiptDate: currentMonthStart.AddDate(0,0,1).Format("2006-01-02")}, // 5000 total for acc1 current month
		{BankAccountID: "acc2", Amount: 1000.00, ReceiptDate: currentMonthStart.AddDate(0, -1, 5).Format("2006-01-02")}, // Previous month
	}
	// We expect GetIncomeRecords to be called with UserID "test-user-123"
	mockIncomeSvc.On("GetIncomeRecords", ctx, &financeEntity.GetIncomeRecordsQueryParameters{UserID: "test-user-123"}).Return(mockIncomes, nil)

	// --- Mocking ExpenseRecordService ---
	// Paid expense from acc1 (current month)
	// Paid expense from acc1 (previous month)
	// Unpaid expense (current month, upcoming bill)
	// Paid expense from acc2 (current month)
	paidDateCurrentMonth := currentMonthStart.AddDate(0,0,2).Format("2006-01-02")
	paidDatePreviousMonth := currentMonthStart.AddDate(0,-1,3).Format("2006-01-02")
	unpaidDueDate := currentMonthStart.AddDate(0,0,15).Format("2006-01-02")
	descUnpaid := "Aluguel Mensal" // Example description for unpaid bill

	mockRawExpenses := []financeEntity.ExpenseRecord{
		{ID: "exp1", BankPaidFrom: &mockAccounts[0].ID, Amount: 500.00, PaymentDate: &paidDateCurrentMonth, Category: "Alimentação", DueDate: currentMonthStart.Format("2006-01-02")},
		{ID: "exp2", BankPaidFrom: &mockAccounts[0].ID, Amount: 200.00, PaymentDate: &paidDatePreviousMonth, Category: "Transporte", DueDate: currentMonthStart.AddDate(0,-1,1).Format("2006-01-02")},
		{ID: "exp3", UserID: "test-user-123", Amount: 1500.00, PaymentDate: nil, Category: "Moradia", DueDate: unpaidDueDate, Description: &descUnpaid},
		{ID: "exp4", BankPaidFrom: &mockAccounts[1].ID, Amount: 300.00, PaymentDate: &paidDateCurrentMonth, Category: "Lazer", DueDate: currentMonthStart.Format("2006-01-02")},
	}
	mockExpenseSvc.On("GetExpenseRecords", ctx).Return(mockRawExpenses, nil)


	// --- Mocking ProfileGoalsService ---
	mockProfileGoals := profileEntity.ProfileGoals{
		Goals2Years: []profileEntity.Goals{
			{Name: "Viagem", TargetAmount: 5000}, // Assume not completed
			{Name: "Curso", TargetAmount: 1000},  // Assume not completed
		},
		Goals5Years: []profileEntity.Goals{
			{Name: "Carro", TargetAmount: 20000}, // Assume not completed
		},
	}
	userIDStr := "test-user-123"
	mockGoalsSvc.On("GetProfileGoals", ctx, &userIDStr).Return(mockProfileGoals, nil)

	// --- Call the method under test ---
	dashboard, err := dashboardService.GetDashboardData(ctx)

	// --- Assertions ---
	assert.NoError(t, err)
	assert.NotNil(t, dashboard)

	// Summary Cards
	// Acc1: Income 5000 (current) - Expense 500 (current) = 4500
	// Acc2: Income 0 (current) - Expense 300 (current) = -300
	// Total Balance = 4500 - 300 = 4200
	assert.Equal(t, 4200.00, dashboard.SummaryCards.TotalBalance)
	assert.Equal(t, 5000.00, dashboard.SummaryCards.MonthlyRevenue)  // 2000 + 3000 from acc1
	assert.Equal(t, 800.00, dashboard.SummaryCards.MonthlyExpenses) // 500 from acc1 + 300 from acc2
	assert.Equal(t, "0% (0 de 3 metas)", dashboard.SummaryCards.GoalsProgress) // Due to limitation

	// Account Summary
	assert.Len(t, dashboard.AccountSummaryData, 2)
	for _, as := range dashboard.AccountSummaryData {
		if as.AccountName == "Conta Corrente A" {
			assert.Equal(t, 4500.00, as.Balance)
		} else if as.AccountName == "Poupança B" {
			assert.Equal(t, -300.00, as.Balance)
		} else {
			t.Errorf("Unexpected account summary name: %s", as.AccountName)
		}
	}

	// Upcoming Bills
	assert.Len(t, dashboard.UpcomingBillsData, 1)
	if len(dashboard.UpcomingBillsData) > 0 {
		assert.Equal(t, "Aluguel Mensal", dashboard.UpcomingBillsData[0].BillName) // Was Moradia, changed to use Description
		assert.Equal(t, 1500.00, dashboard.UpcomingBillsData[0].Amount)
		parsedDueDate, _ := time.Parse("2006-01-02", unpaidDueDate)
		assert.Equal(t, parsedDueDate, dashboard.UpcomingBillsData[0].DueDate)
	}

	// RevenueExpenseChartData (check for 6 months, values for current month)
	assert.Len(t, dashboard.RevenueExpenseChartData, 6)
    currentMonthChartItem := dashboard.RevenueExpenseChartData[5] // Last item is current month
    assert.Equal(t, currentMonthStart.Format("Jan/06"), currentMonthChartItem.Month)
    assert.Equal(t, 5000.00, currentMonthChartItem.Revenue)
    assert.Equal(t, 800.00, currentMonthChartItem.Expenses)


	// ExpenseCategoryChartData
	assert.Len(t, dashboard.ExpenseCategoryChartData, 2) // Alimentação, Lazer for current month
    foundAlimentacao := false
    foundLazer := false
    for _, cat := range dashboard.ExpenseCategoryChartData {
        if cat.Name == "Alimentação" {
            assert.Equal(t, 500.00, cat.Value)
            foundAlimentacao = true
        } else if cat.Name == "Lazer" {
            assert.Equal(t, 300.00, cat.Value)
            foundLazer = true
        } else {
            t.Errorf("Unexpected expense category: %s", cat.Name)
        }
    }
    assert.True(t, foundAlimentacao, "Alimentação category not found or value incorrect")
    assert.True(t, foundLazer, "Lazer category not found or value incorrect")


	// Verify that all expected mock calls were made
	mockBankSvc.AssertExpectations(t)
	mockExpenseSvc.AssertExpectations(t)
	mockIncomeSvc.AssertExpectations(t)
	mockGoalsSvc.AssertExpectations(t)
}

// TODO: Add more test cases:
// - Test with no bank accounts
// - Test with no income/expenses
// - Test with no goals
// - Test when a dependent service returns an error (e.g., mockBankSvc.On("GetBankAccounts", ctx).Return(nil, fmt.Errorf("db error")))
// - Test edge cases for date calculations (e.g., start/end of year for charts)
