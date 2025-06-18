package dashboard

import (
	"context"
	"fmt"
	"log"
	"sort"

	// "strconv" // Was potentially for GoalsProgress, check if still needed
	"time"

	dashboardEntity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
	financeEntity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	profileGoals "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
	profileEntity "github.com/Tomelin/dashfin-backend-app/internal/core/service/profile"
)

const defaultDashboardCacheTTL = 5 * time.Minute // Example TTL for dashboard cache

// DashboardService provides the logic for aggregating dashboard data.
type DashboardService struct {
	bankAccountService   financeEntity.BankAccountServiceInterface
	expenseRecordService financeEntity.ExpenseRecordServiceInterface
	incomeRecordService  financeEntity.IncomeRecordServiceInterface
	profileGoalsService  profileEntity.ProfileGoalsServiceInterface
	dashboardRepository  dashboardEntity.DashboardRepositoryInterface // New dependency
}

// NewDashboardService creates a new DashboardService.
func NewDashboardService(
	bankAccountSvc financeEntity.BankAccountServiceInterface,
	expenseRecordSvc financeEntity.ExpenseRecordServiceInterface,
	incomeRecordSvc financeEntity.IncomeRecordServiceInterface,
	profileGoalsSvc profileEntity.ProfileGoalsServiceInterface,
	dashboardRepo dashboardEntity.DashboardRepositoryInterface, // New dependency
) *DashboardService {
	return &DashboardService{
		bankAccountService:   bankAccountSvc,
		expenseRecordService: expenseRecordSvc,
		incomeRecordService:  incomeRecordSvc,
		profileGoalsService:  profileGoalsSvc,
		dashboardRepository:  dashboardRepo, // Store the new dependency
	}
}

// GetDashboardData aggregates all necessary data for the financial dashboard.
// It now includes caching logic.
func (s *DashboardService) GetDashboardData(ctx context.Context) (*dashboardEntity.Dashboard, error) {
	userIDFromCtx := ctx.Value("UserID")
	if userIDFromCtx == nil {
		return nil, fmt.Errorf("userID not found in context")
	}
	userID, ok := userIDFromCtx.(string)
	if !ok {
		return nil, fmt.Errorf("userID in context is not a string")
	}
	if userID == "" {
		return nil, fmt.Errorf("userID in context is empty")
	}

	// 1. Try to fetch from cache first
	cachedDashboard, found, err := s.dashboardRepository.GetDashboard(ctx, userID)
	if err != nil {
		// Log error but proceed to generate data, cache error shouldn't break main functionality
		fmt.Printf("Warning: Error fetching dashboard from cache for user %s: %v\n", userID, err)
	}
	if found && cachedDashboard != nil {
		log.Println("Dashboard found in cache", cachedDashboard, found)
		// return cachedDashboard, nil
	}

	// 2. If not in cache or error during cache fetch, generate fresh data
	dashboard, err := s.generateFreshDashboardData(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("generating fresh dashboard data: %w", err)
	}

	balanceCard := []dashboardEntity.AccountBalanceItem{
		{
			ID:          "0001",
			AccountName: "Itau",
			BankName:    "Itau",
			Balance:     78232.00,
		},
		{
			ID:          "0002",
			AccountName: "Santander",
			BankName:    "Santander",
			Balance:     43.00,
		},
		{
			ID:          "0003",
			AccountName: "Caixa",
			BankName:    "Caixa",
			Balance:     782.00,
		},
	}

	monthlyFinancial := []dashboardEntity.MonthlyFinancialSummaryItem{
		{
			Month:         "2025-05",
			TotalIncome:   5123.98,
			TotalExpenses: 2345.00,
		},
		{
			Month:         "2025-04",
			TotalIncome:   6789.98,
			TotalExpenses: 7865.75,
		},
		{
			Month:         "2025-05",
			TotalIncome:   9856.98,
			TotalExpenses: 8764.00,
		},
		{
			Month:         "2025-02",
			TotalIncome:   9876.98,
			TotalExpenses: 2345.00,
		}}

	dashboard.SummaryCards.AccountBalances = balanceCard
	dashboard.SummaryCards.MonthlyFinancialSummary = monthlyFinancial

	// 3. Save the newly generated dashboard to cache
	// Use a default TTL, this could be configurable
	err = s.dashboardRepository.SaveDashboard(ctx, userID, dashboard, defaultDashboardCacheTTL)
	if err != nil {
		// Log error but don't fail the main operation, dashboard was generated successfully
		fmt.Printf("Warning: Error saving dashboard to cache for user %s: %v\n", userID, err)
	}

	return dashboard, nil
}

// generateFreshDashboardData contains the original logic to build the dashboard from various services.
func (s *DashboardService) generateFreshDashboardData(ctx context.Context, userID string) (*dashboardEntity.Dashboard, error) {
	now := time.Now()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	allUserBankAccounts, err := s.bankAccountService.GetBankAccounts(ctx)
	if err != nil {
		fmt.Printf("Warning: Error fetching bank accounts for user %s: %v\n", userID, err)
		allUserBankAccounts = []financeEntity.BankAccountRequest{}
	}

	allUserIncomes, err := s.incomeRecordService.GetIncomeRecords(ctx, &financeEntity.GetIncomeRecordsQueryParameters{UserID: userID})
	if err != nil {
		fmt.Printf("Warning: Error fetching income records for user %s: %v\n", userID, err)
		allUserIncomes = []financeEntity.IncomeRecord{}
	}

	allUserRawExpenses, err := s.expenseRecordService.GetExpenseRecords(ctx)
	if err != nil {
		fmt.Printf("Warning: Error fetching expense records for user %s: %v\n", userID, err)
		allUserRawExpenses = []financeEntity.ExpenseRecord{}
	}

	allUserPaidExpenses := make([]financeEntity.ExpenseRecord, 0)
	for _, exp := range allUserRawExpenses {
		if exp.PaymentDate != nil && *exp.PaymentDate != "" {
			paymentT, parseErr := time.Parse("2006-01-02", *exp.PaymentDate)
			if parseErr == nil && !paymentT.After(now) {
				allUserPaidExpenses = append(allUserPaidExpenses, exp)
			}
		}
	}

	totalBalance := s.calculateTotalBalance(allUserBankAccounts, allUserIncomes, allUserPaidExpenses)
	monthlyRevenue := s.calculateMonthlyRevenue(allUserIncomes, currentMonthStart)
	monthlyExpenses := s.calculateMonthlyExpenses(allUserPaidExpenses, currentMonthStart)

	goalsProgressStr := "N/A (Data unavailable)"
	profileGoals, err := s.profileGoalsService.GetProfileGoals(ctx, &userID)

	if err != nil {
		fmt.Printf("Warning: Error fetching profile goals for user %s: %v\n", userID, err)
	} else {
		goalsProgressStr = s.formatGoalsProgress(profileGoals)
	}

	dashboard := &dashboardEntity.Dashboard{
		SummaryCards: dashboardEntity.SummaryCards{
			TotalBalance:    totalBalance,
			MonthlyRevenue:  monthlyRevenue,
			MonthlyExpenses: monthlyExpenses,
			GoalsProgress:   goalsProgressStr,
		},
	}

	dashboard.AccountSummaryData = s.getAccountSummaries(allUserBankAccounts, allUserIncomes, allUserPaidExpenses)
	upcomingBills, err := s.getUpcomingBills(ctx, userID, now, allUserRawExpenses)
	if err != nil {
		fmt.Printf("Warning: Error fetching upcoming bills: %v\n", err)
		dashboard.UpcomingBillsData = []dashboardEntity.UpcomingBill{}
	} else {
		dashboard.UpcomingBillsData = upcomingBills
	}

	dashboard.RevenueExpenseChartData = s.getRevenueExpenseChartData(allUserIncomes, allUserPaidExpenses, now, 6)
	expenseCategories, err := s.getExpenseCategoriesForMonth(allUserPaidExpenses, currentMonthStart)
	if err != nil {
		fmt.Printf("Warning: Error fetching expense categories: %v\n", err)
		dashboard.ExpenseCategoryChartData = []dashboardEntity.ExpenseCategoryChartItem{}
	} else {
		dashboard.ExpenseCategoryChartData = expenseCategories
	}

	dashboard.PersonalizedRecommendationsData = []dashboardEntity.PersonalizedRecommendation{}

	return dashboard, nil
}

// --- Helper methods (calculateTotalBalance, etc.) remain the same as before ---
// Ensure they are part of the DashboardService (s *DashboardService)

func (s *DashboardService) calculateTotalBalance(
	accounts []financeEntity.BankAccountRequest,
	incomes []financeEntity.IncomeRecord,
	paidExpenses []financeEntity.ExpenseRecord,
) float64 {
	var totalBalance float64
	accountBalances := s.calculateAllAccountBalances(accounts, incomes, paidExpenses)
	for _, balance := range accountBalances {
		totalBalance += balance
	}
	return totalBalance
}

func (s *DashboardService) calculateAllAccountBalances(
	accounts []financeEntity.BankAccountRequest,
	incomes []financeEntity.IncomeRecord,
	paidExpenses []financeEntity.ExpenseRecord,
) map[string]float64 {
	accountBalances := make(map[string]float64)
	for _, acc := range accounts {
		accountBalances[acc.ID] = 0
	}
	for _, income := range incomes {
		accountBalances[income.BankAccountID] += income.Amount
	}
	for _, expense := range paidExpenses {
		if expense.BankPaidFrom != nil && *expense.BankPaidFrom != "" {
			accountBalances[*expense.BankPaidFrom] -= expense.Amount
		}
	}
	return accountBalances
}

func (s *DashboardService) calculateMonthlyRevenue(
	incomes []financeEntity.IncomeRecord,
	monthStart time.Time,
) float64 {
	var totalRevenue float64
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	for _, income := range incomes {
		receiptDate, err := time.Parse("2006-01-02", income.ReceiptDate)
		if err == nil {
			if !receiptDate.Before(monthStart) && !receiptDate.After(monthEnd) {
				totalRevenue += income.Amount
			}
		}
	}
	return totalRevenue
}

func (s *DashboardService) calculateMonthlyExpenses(
	paidExpenses []financeEntity.ExpenseRecord,
	monthStart time.Time,
) float64 {
	var totalExpenses float64
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	for _, expense := range paidExpenses {
		if expense.PaymentDate != nil && *expense.PaymentDate != "" {
			paymentDate, err := time.Parse("2006-01-02", *expense.PaymentDate)
			if err == nil {
				if !paymentDate.Before(monthStart) && !paymentDate.After(monthEnd) {
					totalExpenses += expense.Amount
				}
			}
		}
	}
	return totalExpenses
}

func (s *DashboardService) getAccountSummaries(
	accounts []financeEntity.BankAccountRequest,
	incomes []financeEntity.IncomeRecord,
	paidExpenses []financeEntity.ExpenseRecord,
) []dashboardEntity.AccountSummary {
	accountBalances := s.calculateAllAccountBalances(accounts, incomes, paidExpenses)
	summaries := make([]dashboardEntity.AccountSummary, 0, len(accounts))
	for _, acc := range accounts {
		summaryName := acc.CustomBankName
		if summaryName == "" {
			summaryName = acc.BankCode + " - " + acc.AccountNumber
		}
		summaries = append(summaries, dashboardEntity.AccountSummary{
			AccountName: summaryName,
			Balance:     accountBalances[acc.ID],
		})
	}
	return summaries
}

func (s *DashboardService) formatGoalsProgress(profileGoals profileGoals.ProfileGoals) string {
	allGoals := append(profileGoals.Goals2Years, profileGoals.Goals5Years...)
	allGoals = append(allGoals, profileGoals.Goals10Years...) // Typo: allGolas -> allGoals
	totalGoals := len(allGoals)
	completedGoals := 0 // Limitation: Cannot determine completed goals
	if totalGoals == 0 {
		return "Nenhuma meta definida"
	}
	percentage := 0.0
	return fmt.Sprintf("%.0f%% (%d de %d metas)", percentage, completedGoals, totalGoals)
}

func (s *DashboardService) getUpcomingBills(
	ctx context.Context,
	userID string,
	fromDate time.Time,
	allUserRawExpenses []financeEntity.ExpenseRecord,
) ([]dashboardEntity.UpcomingBill, error) {
	upcomingEndDate := fromDate.AddDate(0, 0, 30)
	upcomingStartDate := fromDate.AddDate(0, 0, -60)

	bills := make([]dashboardEntity.UpcomingBill, 0)
	for _, exp := range allUserRawExpenses {
		if exp.PaymentDate == nil || *exp.PaymentDate == "" {
			dueDate, errParse := time.Parse("2006-01-02", exp.DueDate)
			if errParse != nil {
				fmt.Printf("Warning: Could not parse DueDate '%s' for expense ID %s: %v\n", exp.DueDate, exp.ID, errParse)
				continue
			}
			if !dueDate.Before(upcomingStartDate) && !dueDate.After(upcomingEndDate) {
				if exp.PaymentDate == nil || *exp.PaymentDate == "" {
					billName := exp.Description
					if billName == nil || *billName == "" {
						billName = &exp.Category
					}
					bills = append(bills, dashboardEntity.UpcomingBill{
						BillName: *billName,
						Amount:   exp.Amount,
						DueDate:  dueDate.Format("2006-01-02"), // Assign the parsed time.Time value
					})
				}
			}
		}
	}
	sort.Slice(bills, func(i, j int) bool {
		// Parse DueDate strings back to time.Time for comparison
		dateI, _ := time.Parse("2006-01-02", bills[i].DueDate)
		dateJ, _ := time.Parse("2006-01-02", bills[j].DueDate)

		// Handle parsing errors, maybe consider invalid dates as later or earlier depending on desired sort behavior
		// For simplicity, if one fails, the order might be unpredictable for that pair.
		// A more robust solution would handle errors or ensure dates are always parseable.
		return dateI.Before(dateJ)
	})
	return bills, nil
}

func (s *DashboardService) getRevenueExpenseChartData(
	incomes []financeEntity.IncomeRecord,
	paidExpenses []financeEntity.ExpenseRecord,
	currentDate time.Time,
	numberOfMonths int,
) []dashboardEntity.RevenueExpenseChartItem {
	chartData := make([]dashboardEntity.RevenueExpenseChartItem, numberOfMonths)
	for i := 0; i < numberOfMonths; i++ {
		monthStart := time.Date(currentDate.Year(), currentDate.Month()-time.Month(i), 1, 0, 0, 0, 0, time.UTC)
		monthlyRevenue := s.calculateMonthlyRevenue(incomes, monthStart)
		monthlyExpenses := s.calculateMonthlyExpenses(paidExpenses, monthStart)
		chartData[numberOfMonths-1-i] = dashboardEntity.RevenueExpenseChartItem{
			Month:    monthStart.Format("Jan/06"),
			Revenue:  monthlyRevenue,
			Expenses: monthlyExpenses,
		}
	}
	return chartData
}

func (s *DashboardService) getExpenseCategoriesForMonth(
	paidExpenses []financeEntity.ExpenseRecord,
	monthStart time.Time,
) ([]dashboardEntity.ExpenseCategoryChartItem, error) {
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	categories := make(map[string]float64)
	for _, exp := range paidExpenses {
		if exp.PaymentDate != nil && *exp.PaymentDate != "" {
			paymentDate, err := time.Parse("2006-01-02", *exp.PaymentDate)
			if err == nil {
				if !paymentDate.Before(monthStart) && !paymentDate.After(monthEnd) {
					categoryName := exp.Category
					if categoryName == "" {
						categoryName = "Outros"
					}
					categories[categoryName] += exp.Amount
				}
			}
		}
	}
	chartData := make([]dashboardEntity.ExpenseCategoryChartItem, 0, len(categories))
	for name, value := range categories {
		chartData = append(chartData, dashboardEntity.ExpenseCategoryChartItem{
			Name:  name,
			Value: value,
		})
	}
	sort.Slice(chartData, func(i, j int) bool {
		return chartData[i].Value > chartData[j].Value
	})
	return chartData, nil
}
