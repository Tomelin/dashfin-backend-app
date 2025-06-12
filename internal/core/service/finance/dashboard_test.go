package finance

import (
	"context"
	"errors" // For creating test errors
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	financeEntity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	profileEntity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
)

// --- Existing Mocks (BankAccountService, ExpenseRecordService, IncomeRecordService, ProfileGoalsService) remain the same ---

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
	args := m.Called(ctx, userID)
	return args.Get(0).(profileEntity.ProfileGoals), args.Error(1)
}


// --- New Mock for DashboardRepositoryInterface ---
type MockDashboardRepository struct {
	mock.Mock
}

func (m *MockDashboardRepository) GetDashboard(ctx context.Context, userID string) (*financeEntity.Dashboard, bool, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*financeEntity.Dashboard), args.Bool(1), args.Error(2)
}

func (m *MockDashboardRepository) SaveDashboard(ctx context.Context, userID string, dashboard *financeEntity.Dashboard, ttl time.Duration) error {
	args := m.Called(ctx, userID, dashboard, ttl)
	return args.Error(0)
}

func (m *MockDashboardRepository) DeleteDashboard(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}


// TestDashboardService_GetDashboardData_CacheMiss (Refactored from original Nominal test)
func TestDashboardService_GetDashboardData_CacheMiss(t *testing.T) {
	mockBankSvc := new(MockBankAccountService)
	mockExpenseSvc := new(MockExpenseRecordService)
	mockIncomeSvc := new(MockIncomeRecordService)
	mockGoalsSvc := new(MockProfileGoalsService)
	mockDashboardRepo := new(MockDashboardRepository) // New mock

	// Pass all mocks to the constructor
	dashboardService := NewDashboardService(mockBankSvc, mockExpenseSvc, mockIncomeSvc, mockGoalsSvc, mockDashboardRepo)

	ctx := context.WithValue(context.Background(), "UserID", "test-user-123")
	now := time.Now()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	userIDStr := "test-user-123"

	// --- Mocking DashboardRepository: Cache Miss ---
	mockDashboardRepo.On("GetDashboard", ctx, userIDStr).Return(nil, false, nil).Once()
	// Expect SaveDashboard to be called
	mockDashboardRepo.On("SaveDashboard", ctx, userIDStr, mock.AnythingOfType("*finance.Dashboard"), defaultDashboardCacheTTL).Return(nil).Once()


	// --- Mocking other services (as data will be generated) ---
	mockAccounts := []financeEntity.BankAccountRequest{
		{ID: "acc1", BankAccount: financeEntity.BankAccount{CustomBankName: "Conta Corrente A", AccountNumber: "111"}},
		{ID: "acc2", BankAccount: financeEntity.BankAccount{CustomBankName: "Poupança B", AccountNumber: "222"}},
	}
	mockBankSvc.On("GetBankAccounts", ctx).Return(mockAccounts, nil).Once()

	mockIncomes := []financeEntity.IncomeRecord{
		{BankAccountID: "acc1", Amount: 2000.00, ReceiptDate: currentMonthStart.Format("2006-01-02")},
		{BankAccountID: "acc1", Amount: 3000.00, ReceiptDate: currentMonthStart.AddDate(0,0,1).Format("2006-01-02")},
		{BankAccountID: "acc2", Amount: 1000.00, ReceiptDate: currentMonthStart.AddDate(0, -1, 5).Format("2006-01-02")},
	}
	mockIncomeSvc.On("GetIncomeRecords", ctx, &financeEntity.GetIncomeRecordsQueryParameters{UserID: userIDStr}).Return(mockIncomes, nil).Once()

	paidDateCurrentMonth := currentMonthStart.AddDate(0,0,2).Format("2006-01-02")
	paidDatePreviousMonth := currentMonthStart.AddDate(0,-1,3).Format("2006-01-02")
	unpaidDueDate := currentMonthStart.AddDate(0,0,15).Format("2006-01-02")
	descUnpaid := "Aluguel Mensal"
	mockRawExpenses := []financeEntity.ExpenseRecord{
		{ID: "exp1", BankPaidFrom: &mockAccounts[0].ID, Amount: 500.00, PaymentDate: &paidDateCurrentMonth, Category: "Alimentação", DueDate: currentMonthStart.Format("2006-01-02")},
		{ID: "exp2", BankPaidFrom: &mockAccounts[0].ID, Amount: 200.00, PaymentDate: &paidDatePreviousMonth, Category: "Transporte", DueDate: currentMonthStart.AddDate(0,-1,1).Format("2006-01-02")},
		{ID: "exp3", UserID: userIDStr, Amount: 1500.00, PaymentDate: nil, Category: "Moradia", DueDate: unpaidDueDate, Description: &descUnpaid},
		{ID: "exp4", BankPaidFrom: &mockAccounts[1].ID, Amount: 300.00, PaymentDate: &paidDateCurrentMonth, Category: "Lazer", DueDate: currentMonthStart.Format("2006-01-02")},
	}
	mockExpenseSvc.On("GetExpenseRecords", ctx).Return(mockRawExpenses, nil).Once()

	mockProfileGoals := profileEntity.ProfileGoals{
		Goals2Years: []profileEntity.Goals{{Name: "Viagem", TargetAmount: 5000}, {Name: "Curso", TargetAmount: 1000}},
		Goals5Years: []profileEntity.Goals{{Name: "Carro", TargetAmount: 20000}},
	}
	mockGoalsSvc.On("GetProfileGoals", ctx, &userIDStr).Return(mockProfileGoals, nil).Once()

	// --- Call the method under test ---
	dashboard, err := dashboardService.GetDashboardData(ctx)

	// --- Assertions (same as original nominal test) ---
	assert.NoError(t, err)
	assert.NotNil(t, dashboard)
	assert.Equal(t, 4200.00, dashboard.SummaryCards.TotalBalance)
	assert.Equal(t, 5000.00, dashboard.SummaryCards.MonthlyRevenue)
	assert.Equal(t, 800.00, dashboard.SummaryCards.MonthlyExpenses)
	assert.Equal(t, "0% (0 de 3 metas)", dashboard.SummaryCards.GoalsProgress)

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

	assert.Len(t, dashboard.UpcomingBillsData, 1)
	if len(dashboard.UpcomingBillsData) > 0 {
		assert.Equal(t, descUnpaid, dashboard.UpcomingBillsData[0].BillName)
		assert.Equal(t, 1500.00, dashboard.UpcomingBillsData[0].Amount)
		parsedDueDate, _ := time.Parse("2006-01-02", unpaidDueDate)
		assert.Equal(t, parsedDueDate, dashboard.UpcomingBillsData[0].DueDate)
	}

	assert.Len(t, dashboard.RevenueExpenseChartData, 6)
    currentMonthChartItem := dashboard.RevenueExpenseChartData[5]
    assert.Equal(t, currentMonthStart.Format("Jan/06"), currentMonthChartItem.Month)
    assert.Equal(t, 5000.00, currentMonthChartItem.Revenue)
    assert.Equal(t, 800.00, currentMonthChartItem.Expenses)

	assert.Len(t, dashboard.ExpenseCategoryChartData, 2)
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
    assert.True(t, foundAlimentacao)
    assert.True(t, foundLazer)


	// Verify that all expected mock calls were made
	mockDashboardRepo.AssertExpectations(t)
	mockBankSvc.AssertExpectations(t)
	mockExpenseSvc.AssertExpectations(t)
	mockIncomeSvc.AssertExpectations(t)
	mockGoalsSvc.AssertExpectations(t)
}


func TestDashboardService_GetDashboardData_CacheHit(t *testing.T) {
	mockBankSvc := new(MockBankAccountService)
	mockExpenseSvc := new(MockExpenseRecordService)
	mockIncomeSvc := new(MockIncomeRecordService)
	mockGoalsSvc := new(MockProfileGoalsService)
	mockDashboardRepo := new(MockDashboardRepository)

	dashboardService := NewDashboardService(mockBankSvc, mockExpenseSvc, mockIncomeSvc, mockGoalsSvc, mockDashboardRepo)

	ctx := context.WithValue(context.Background(), "UserID", "test-user-cache-hit")
	userIDStr := "test-user-cache-hit"

	// --- Mocking DashboardRepository: Cache Hit ---
	cachedDashboard := &financeEntity.Dashboard{
		SummaryCards: financeEntity.SummaryCards{TotalBalance: 9999.99, MonthlyRevenue: 100, MonthlyExpenses: 50, GoalsProgress: "Cached"},
		// Populate other fields if necessary for more specific assertions later
	}
	mockDashboardRepo.On("GetDashboard", ctx, userIDStr).Return(cachedDashboard, true, nil).Once()
	// SaveDashboard should NOT be called
	// Other service mocks (Bank, Expense, Income, Goals) should NOT be called

	// --- Call the method under test ---
	dashboard, err := dashboardService.GetDashboardData(ctx)

	// --- Assertions ---
	assert.NoError(t, err)
	assert.NotNil(t, dashboard)
	// Ensure the returned dashboard is the one from the cache
	assert.Equal(t, cachedDashboard.SummaryCards.TotalBalance, dashboard.SummaryCards.TotalBalance)
	assert.Equal(t, "Cached", dashboard.SummaryCards.GoalsProgress)
	assert.Same(t, cachedDashboard, dashboard, "Should return the exact cached dashboard instance")


	// Verify that only GetDashboard was called on the repo, and no other services were hit
	mockDashboardRepo.AssertExpectations(t)
	mockBankSvc.AssertNotCalled(t, "GetBankAccounts", mock.Anything)
	mockExpenseSvc.AssertNotCalled(t, "GetExpenseRecords", mock.Anything)
	mockIncomeSvc.AssertNotCalled(t, "GetIncomeRecords", mock.Anything, mock.Anything)
	mockGoalsSvc.AssertNotCalled(t, "GetProfileGoals", mock.Anything, mock.Anything)
}

func TestDashboardService_GetDashboardData_CacheGetError(t *testing.T) {
    // Similar to CacheMiss, but GetDashboard returns an error.
    // The service should proceed to generate data and attempt to save it.
	mockBankSvc := new(MockBankAccountService)
	mockExpenseSvc := new(MockExpenseRecordService)
	mockIncomeSvc := new(MockIncomeRecordService)
	mockGoalsSvc := new(MockProfileGoalsService)
	mockDashboardRepo := new(MockDashboardRepository)

	dashboardService := NewDashboardService(mockBankSvc, mockExpenseSvc, mockIncomeSvc, mockGoalsSvc, mockDashboardRepo)
	ctx := context.WithValue(context.Background(), "UserID", "test-user-cache-get-err")
	userIDStr := "test-user-cache-get-err"

	// --- Mocking DashboardRepository: Cache Get Error ---
	mockDashboardRepo.On("GetDashboard", ctx, userIDStr).Return(nil, false, errors.New("cache read error")).Once()
	// Expect SaveDashboard to be called as data will be regenerated
	mockDashboardRepo.On("SaveDashboard", ctx, userIDStr, mock.AnythingOfType("*finance.Dashboard"), defaultDashboardCacheTTL).Return(nil).Once()

    // Mocks for data generation services (Bank, Income, Expense, Goals) - simplified for brevity
    mockBankSvc.On("GetBankAccounts", ctx).Return([]financeEntity.BankAccountRequest{}, nil).Once()
    mockIncomeSvc.On("GetIncomeRecords", ctx, &financeEntity.GetIncomeRecordsQueryParameters{UserID: userIDStr}).Return([]financeEntity.IncomeRecord{}, nil).Once()
    mockExpenseSvc.On("GetExpenseRecords", ctx).Return([]financeEntity.ExpenseRecord{}, nil).Once()
    mockGoalsSvc.On("GetProfileGoals", ctx, &userIDStr).Return(profileEntity.ProfileGoals{}, nil).Once()


	dashboard, err := dashboardService.GetDashboardData(ctx)
	assert.NoError(t, err) // Main operation should succeed
	assert.NotNil(t, dashboard)
    // Further assertions on generated data if needed

	mockDashboardRepo.AssertExpectations(t)
	mockBankSvc.AssertExpectations(t)
	mockExpenseSvc.AssertExpectations(t)
	mockIncomeSvc.AssertExpectations(t)
	mockGoalsSvc.AssertExpectations(t)
}


func TestDashboardService_GetDashboardData_CacheSaveError(t *testing.T) {
    // Cache miss, data generated, but SaveDashboard returns an error.
    // Main operation should still succeed.
	mockBankSvc := new(MockBankAccountService)
	mockExpenseSvc := new(MockExpenseRecordService)
	mockIncomeSvc := new(MockIncomeRecordService)
	mockGoalsSvc := new(MockProfileGoalsService)
	mockDashboardRepo := new(MockDashboardRepository)

	dashboardService := NewDashboardService(mockBankSvc, mockExpenseSvc, mockIncomeSvc, mockGoalsSvc, mockDashboardRepo)
	ctx := context.WithValue(context.Background(), "UserID", "test-user-cache-save-err")
	userIDStr := "test-user-cache-save-err"

	// --- Mocking DashboardRepository: Cache Miss, then Save Error ---
	mockDashboardRepo.On("GetDashboard", ctx, userIDStr).Return(nil, false, nil).Once()
	mockDashboardRepo.On("SaveDashboard", ctx, userIDStr, mock.AnythingOfType("*finance.Dashboard"), defaultDashboardCacheTTL).Return(errors.New("cache write error")).Once()

    // Mocks for data generation services (Bank, Income, Expense, Goals) - simplified
    mockBankSvc.On("GetBankAccounts", ctx).Return([]financeEntity.BankAccountRequest{}, nil).Once()
    mockIncomeSvc.On("GetIncomeRecords", ctx, &financeEntity.GetIncomeRecordsQueryParameters{UserID: userIDStr}).Return([]financeEntity.IncomeRecord{}, nil).Once()
    mockExpenseSvc.On("GetExpenseRecords", ctx).Return([]financeEntity.ExpenseRecord{}, nil).Once()
    mockGoalsSvc.On("GetProfileGoals", ctx, &userIDStr).Return(profileEntity.ProfileGoals{}, nil).Once()

	dashboard, err := dashboardService.GetDashboardData(ctx)
	assert.NoError(t, err) // Main operation should succeed
	assert.NotNil(t, dashboard)
    // Further assertions on generated data

	mockDashboardRepo.AssertExpectations(t)
    // Other services also expected to be called
	mockBankSvc.AssertExpectations(t)
	mockExpenseSvc.AssertExpectations(t)
	mockIncomeSvc.AssertExpectations(t)
	mockGoalsSvc.AssertExpectations(t)
}

// Existing tests for other scenarios (no bank accounts, no income/expenses etc.) should also be updated
// to include the DashboardRepository mock, typically expecting a GetDashboard (miss) and SaveDashboard call.
