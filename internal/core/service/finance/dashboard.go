package finance

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	financeEntity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	profileEntity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
	// Actual service interface imports - these paths might need adjustment
	// based on where these interfaces are defined in the project.
	// Assuming they are in the same package as their entities for now.
)

// DashboardService provides the logic for aggregating dashboard data.
type DashboardService struct {
	bankAccountService    financeEntity.BankAccountServiceInterface
	expenseRecordService  financeEntity.ExpenseRecordServiceInterface
	incomeRecordService   financeEntity.IncomeRecordServiceInterface
	profileGoalsService profileEntity.ProfileGoalsServiceInterface // Corrected interface type
}

// NewDashboardService creates a new DashboardService.
func NewDashboardService(
	bankAccountSvc financeEntity.BankAccountServiceInterface,
	expenseRecordSvc financeEntity.ExpenseRecordServiceInterface,
	incomeRecordSvc financeEntity.IncomeRecordServiceInterface,
	profileGoalsSvc profileEntity.ProfileGoalsServiceInterface, // Corrected type
) *DashboardService {
	return &DashboardService{
		bankAccountService:    bankAccountSvc,
		expenseRecordService:  expenseRecordSvc,
		incomeRecordService:   incomeRecordSvc,
		profileGoalsService: profileGoalsSvc,
	}
}

// GetDashboardData aggregates all necessary data for the financial dashboard.
func (s *DashboardService) GetDashboardData(ctx context.Context) (*financeEntity.Dashboard, error) {
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

	now := time.Now()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Fetch all accounts first as they are needed for multiple calculations
	allUserBankAccounts, err := s.bankAccountService.GetBankAccounts(ctx) // Assumes this is user-scoped by context
	if err != nil {
		// If no bank accounts, some parts of dashboard can't be computed.
		// Depending on requirements, might return partial data or error.
		// For now, log and continue where possible.
		fmt.Printf("Warning: Error fetching bank accounts for user %s: %v\n", userID, err)
		// Initialize to empty slice if error, to prevent nil pointer issues later
		allUserBankAccounts = []financeEntity.BankAccountRequest{}
	}

	// Fetch all incomes and expenses for the user (or a relevant recent period)
	// This data will be used for balance calculations and monthly summaries.
	// For simplicity, fetching all records. For performance, might filter by a recent date range.
	allUserIncomes, err := s.incomeRecordService.GetIncomeRecords(ctx, &financeEntity.GetIncomeRecordsQueryParameters{UserID: userID})
	if err != nil {
		fmt.Printf("Warning: Error fetching income records for user %s: %v\n", userID, err)
		allUserIncomes = []financeEntity.IncomeRecord{}
	}

	// Fetch all *paid* expenses. This requires filtering by PaymentDate != nil.
	// The current ExpenseRecordService.GetExpenseRecords might not support this directly.
	// Assuming GetExpenseRecordsByFilter can be used, or GetExpenseRecords and then filter locally.
	// For now, let's fetch all and filter locally.
	// This might be inefficient for users with many expense records.
	allUserRawExpenses, err := s.expenseRecordService.GetExpenseRecords(ctx) // Assumes this is user-scoped
	if err != nil {
		fmt.Printf("Warning: Error fetching expense records for user %s: %v\n", userID, err)
		allUserRawExpenses = []financeEntity.ExpenseRecord{}
	}

	allUserPaidExpenses := make([]financeEntity.ExpenseRecord, 0)
	for _, exp := range allUserRawExpenses {
		if exp.PaymentDate != nil && *exp.PaymentDate != "" {
			// Ensure payment date is in the past or present to be considered "paid" for balance calculation
			paymentT, parseErr := time.Parse("2006-01-02", *exp.PaymentDate)
			if parseErr == nil && !paymentT.After(now) {
				allUserPaidExpenses = append(allUserPaidExpenses, exp)
			}
		}
	}

	// 1. SummaryCards
	totalBalance := s.calculateTotalBalance(allUserBankAccounts, allUserIncomes, allUserPaidExpenses)
	monthlyRevenue := s.calculateMonthlyRevenue(allUserIncomes, currentMonthStart)
	monthlyExpenses := s.calculateMonthlyExpenses(allUserPaidExpenses, currentMonthStart) // Use already filtered paid expenses

	goalsProgressStr := "N/A (Data unavailable)" // Default if goals cannot be processed
	profileGoals, err := s.profileGoalsService.GetProfileGoals(ctx, &userID)
	if err != nil {
		fmt.Printf("Warning: Error fetching profile goals for user %s: %v\n", userID, err)
	} else {
		goalsProgressStr = s.formatGoalsProgress(profileGoals)
	}

	dashboard := &financeEntity.Dashboard{
		SummaryCards: financeEntity.SummaryCards{
			TotalBalance:    totalBalance,
			MonthlyRevenue:  monthlyRevenue,
			MonthlyExpenses: monthlyExpenses,
			GoalsProgress:   goalsProgressStr,
		},
	}

	// 2. AccountSummaryData
	dashboard.AccountSummaryData = s.getAccountSummaries(allUserBankAccounts, allUserIncomes, allUserPaidExpenses)

	// 3. UpcomingBillsData
	upcomingBills, err := s.getUpcomingBills(ctx, userID, now, allUserRawExpenses) // Pass raw expenses
	if err != nil {
		fmt.Printf("Warning: Error fetching upcoming bills: %v\n", err)
		dashboard.UpcomingBillsData = []financeEntity.UpcomingBill{} // Default to empty
	} else {
		dashboard.UpcomingBillsData = upcomingBills
	}

	// 4. RevenueExpenseChartData (e.g., last 6 months)
	dashboard.RevenueExpenseChartData = s.getRevenueExpenseChartData(allUserIncomes, allUserPaidExpenses, now, 6)


	// 5. ExpenseCategoryChartData (current month)
	expenseCategories, err := s.getExpenseCategoriesForMonth(allUserPaidExpenses, currentMonthStart)
	if err != nil {
		fmt.Printf("Warning: Error fetching expense categories: %v\n", err)
		dashboard.ExpenseCategoryChartData = []financeEntity.ExpenseCategoryChartItem{} // Default
	} else {
		dashboard.ExpenseCategoryChartData = expenseCategories
	}

	// 6. PersonalizedRecommendationsData
	dashboard.PersonalizedRecommendationsData = []financeEntity.PersonalizedRecommendation{} // Placeholder

	return dashboard, nil
}

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
        accountBalances[acc.ID] = 0 // Initialize balance for all accounts
    }

    for _, income := range incomes {
        // Assume income.BankAccountID is the ID of the BankAccountRequest
        accountBalances[income.BankAccountID] += income.Amount
    }

    for _, expense := range paidExpenses {
        if expense.BankPaidFrom != nil && *expense.BankPaidFrom != "" {
            // Assume expense.BankPaidFrom stores the ID of the BankAccountRequest
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
	paidExpenses []financeEntity.ExpenseRecord, // Use pre-filtered paid expenses
	monthStart time.Time,
) float64 {
	var totalExpenses float64
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	for _, expense := range paidExpenses {
		// PaymentDate should be used to determine if the expense falls into the current month
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
) []financeEntity.AccountSummary {

	accountBalances := s.calculateAllAccountBalances(accounts, incomes, paidExpenses)
	summaries := make([]financeEntity.AccountSummary, 0, len(accounts))

	for _, acc := range accounts {
		summaryName := acc.CustomBankName
		if summaryName == "" {
			summaryName = acc.BankCode + " - " + acc.AccountNumber // Default name
		}
		summaries = append(summaries, financeEntity.AccountSummary{
			AccountName: summaryName,
			Balance:     accountBalances[acc.ID], // Use calculated balance
		})
	}
	return summaries
}

func (s *DashboardService) formatGoalsProgress(profileGoals profileEntity.ProfileGoals) string {
	allGoals := append(profileGoals.Goals2Years, profileGoals.Goals5Years...)
	allGoals = append(allGoals, profileGoals.Goals10Years...)

	totalGoals := len(allGoals)
	completedGoals := 0

	// CRITICAL LIMITATION: The entity_profile.Goals struct does NOT have a field
	// like 'IsCompleted' or 'CurrentAmount'. So, we cannot accurately determine
	// completed goals or percentage progress.
	// For now, assuming 0 completed until the entity is updated.
	// completedGoals = countCompleted(allGoals) // This function cannot be implemented yet

	if totalGoals == 0 {
		return "Nenhuma meta definida"
	}

	// Placeholder logic due to missing 'IsCompleted' field in Goals entity
	// This will always show 0% and (0 of X goals)
	percentage := 0.0
	if totalGoals > 0 {
		// percentage = (float64(completedGoals) / float64(totalGoals)) * 100 // This would be the ideal calculation
	}

	return fmt.Sprintf("%.0f%% (%d de %d metas)", percentage, completedGoals, totalGoals)
}


func (s *DashboardService) getUpcomingBills(
    ctx context.Context,
    userID string,
    fromDate time.Time,
    allUserRawExpenses []financeEntity.ExpenseRecord, // Pass all raw expenses
) ([]financeEntity.UpcomingBill, error) {
	upcomingEndDate := fromDate.AddDate(0, 0, 30) // Next 30 days
	bills := make([]financeEntity.UpcomingBill, 0)

	for _, exp := range allUserRawExpenses {
		// Bill is upcoming if:
		// 1. It's not yet paid (PaymentDate is nil or empty)
		// 2. DueDate is within the upcoming window (fromDate to upcomingEndDate)
		if exp.PaymentDate == nil || *exp.PaymentDate == "" {
			dueDate, errParse := time.Parse("2006-01-02", exp.DueDate)
			if errParse != nil {
				fmt.Printf("Warning: Could not parse DueDate '%s' for expense ID %s: %v\n", exp.DueDate, exp.ID, errParse)
				continue
			}

			if !dueDate.Before(fromDate) && !dueDate.After(upcomingEndDate) {
				billName := exp.Description
				if billName == nil || *billName == "" {
					billName = &exp.Category // Use category if description is empty
				}
				bills = append(bills, financeEntity.UpcomingBill{
					BillName: *billName,
					Amount:   exp.Amount,
					DueDate:  dueDate,
				})
			}
		}
	}
	// Sort by DueDate
	sort.Slice(bills, func(i, j int) bool {
		return bills[i].DueDate.Before(bills[j].DueDate)
	})
	return bills, nil
}

func (s *DashboardService) getRevenueExpenseChartData(
    incomes []financeEntity.IncomeRecord,
    paidExpenses []financeEntity.ExpenseRecord,
    currentDate time.Time,
    numberOfMonths int,
) []financeEntity.RevenueExpenseChartItem {
    chartData := make([]financeEntity.RevenueExpenseChartItem, numberOfMonths)

    for i := 0; i < numberOfMonths; i++ {
        // Go back i months from the current month
        monthStart := time.Date(currentDate.Year(), currentDate.Month()-time.Month(i), 1, 0, 0, 0, 0, time.UTC)
        // monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond) // Not strictly needed for current logic

        monthlyRevenue := s.calculateMonthlyRevenue(incomes, monthStart)
        monthlyExpenses := s.calculateMonthlyExpenses(paidExpenses, monthStart)

        chartData[numberOfMonths-1-i] = financeEntity.RevenueExpenseChartItem{ // Store in reverse for correct order
            Month:    monthStart.Format("Jan/06"),
            Revenue:  monthlyRevenue,
            Expenses: monthlyExpenses,
        }
    }
    return chartData
}


func (s *DashboardService) getExpenseCategoriesForMonth(
    paidExpenses []financeEntity.ExpenseRecord, // Use pre-filtered paid expenses
    monthStart time.Time,
) ([]financeEntity.ExpenseCategoryChartItem, error) {
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	categories := make(map[string]float64)

	for _, exp := range paidExpenses {
		if exp.PaymentDate != nil && *exp.PaymentDate != "" {
			paymentDate, err := time.Parse("2006-01-02", *exp.PaymentDate)
			if err == nil {
				if !paymentDate.Before(monthStart) && !paymentDate.After(monthEnd) {
					categoryName := exp.Category
					if categoryName == "" {
						categoryName = "Outros" // Default category
					}
					categories[categoryName] += exp.Amount
				}
			}
		}
	}

	chartData := make([]financeEntity.ExpenseCategoryChartItem, 0, len(categories))
	for name, value := range categories {
		chartData = append(chartData, financeEntity.ExpenseCategoryChartItem{
			Name:  name,
			Value: value,
		})
	}
    // Sort by value descending for better chart display
    sort.Slice(chartData, func(i, j int) bool {
        return chartData[i].Value > chartData[j].Value
    })
	return chartData, nil
}
