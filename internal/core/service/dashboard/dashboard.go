package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	// "strconv" // Was potentially for GoalsProgress, check if still needed
	"time"

	entity_common "github.com/Tomelin/dashfin-backend-app/internal/core/entity/common"
	dashboardEntity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
	financeEntity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	platformInstitution "github.com/Tomelin/dashfin-backend-app/internal/core/entity/platform"
	profileEntity "github.com/Tomelin/dashfin-backend-app/internal/core/service/profile"
	"github.com/Tomelin/dashfin-backend-app/pkg/message_queue"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

const defaultDashboardCacheTTL = 30 * time.Second // Example TTL for dashboard cache

// DashboardService provides the logic for aggregating dashboard data.
type DashboardService struct {
	bankAccountService   financeEntity.BankAccountServiceInterface
	expenseRecordService financeEntity.ExpenseRecordServiceInterface
	incomeRecordService  financeEntity.IncomeRecordServiceInterface
	profileGoalsService  profileEntity.ProfileGoalsServiceInterface
	dashboardRepository  dashboardEntity.DashboardRepositoryInterface // New dependency
	messageQueue         message_queue.MessageQueue
	platformInstitution  platformInstitution.FinancialInstitutionInterface
	dash                 dashboardEntity.Dashboard
	incomeRecords        []financeEntity.IncomeRecord
	expenseRecords       []financeEntity.ExpenseRecord
}

// NewDashboardService creates a new DashboardService.
func NewDashboardService(
	bankAccountSvc financeEntity.BankAccountServiceInterface,
	expenseRecordSvc financeEntity.ExpenseRecordServiceInterface,
	incomeRecordSvc financeEntity.IncomeRecordServiceInterface,
	profileGoalsSvc profileEntity.ProfileGoalsServiceInterface,
	dashboardRepo dashboardEntity.DashboardRepositoryInterface, // New dependency
	messageQueue message_queue.MessageQueue,
	platformInstitution platformInstitution.FinancialInstitutionInterface,
) *DashboardService {

	dash := &DashboardService{
		bankAccountService:   bankAccountSvc,
		expenseRecordService: expenseRecordSvc,
		incomeRecordService:  incomeRecordSvc,
		profileGoalsService:  profileGoalsSvc,
		dashboardRepository:  dashboardRepo, // Store the new dependency
		messageQueue:         messageQueue,
		platformInstitution:  platformInstitution,
	}

	go dash.accountBalance(context.Background())

	return dash
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

	s.dash = dashboardEntity.Dashboard{}

	// 4. Income fetch and set additional data
	err := s.getIncomeRecords(ctx)
	if err != nil {
		log.Println(fmt.Errorf("error fetching income records: %w", err))
	}
	// 5. Expense fetch and set additional data
	err = s.getExpenseRecords(ctx)
	if err != nil {
		log.Println(fmt.Errorf("error fetching expense records: %w", err))
	}

	// 6. Goals fetch and set additional data
	s.formatGoalsProgress(ctx, userID)

	// 7. Get summary cards data
	err = s.getSummaryCards()
	if err != nil {
		log.Println(fmt.Errorf("error getting summary cards: %w", err))
	}

	// 8. Get upcoming bills
	err = s.getUpcomingBills2()
	if err != nil {
		log.Println(fmt.Errorf("error getting upcoming bills: %w", err))
	}

	s.getBankAccountBalance(ctx, &userID)
	// 9. Set the income and expense records
	s.calculateTotalBalance(ctx, userID)

	return &s.dash, nil
}

func (s *DashboardService) getSummaryCards() error {

	var receiveMonth float64
	var expenseMonth float64
	var receiveBalance float64
	var expenseBalance float64

	for _, income := range s.incomeRecords {
		if income.ReceiptDate.After(utils.GetFirstDayOfCurrentMonth()) && income.ReceiptDate.Before(utils.GetLastDayOfCurrentMonth()) {
			receiveMonth += income.Amount
		}
		receiveBalance += income.Amount
	}

	for _, expense := range s.expenseRecords {
		if expense.DueDate.After(utils.GetFirstDayOfCurrentMonth()) && expense.DueDate.Before(utils.GetLastDayOfCurrentMonth()) {
			log.Println("\n Expense record:", expense)
			expenseMonth += expense.Amount
		}
		if expense.DueDate.Before(utils.GetLastDayOfCurrentMonth()) {
			expenseBalance += expense.Amount
		}
	}

	s.dash.SummaryCards.MonthlyExpenses = expenseMonth
	s.dash.SummaryCards.MonthlyRevenue = receiveMonth
	s.dash.SummaryCards.TotalBalance = receiveBalance - expenseBalance

	return nil

}

func (s *DashboardService) getIncomeRecords(ctx context.Context) error {

	records, err := s.incomeRecordService.GetIncomeRecords(ctx, &financeEntity.GetIncomeRecordsQueryParameters{})
	if err != nil {
		return fmt.Errorf("error fetching income records: %w", err)
	}

	s.incomeRecords = records

	return nil
}

func (s *DashboardService) getExpenseRecords(ctx context.Context) error {

	records, err := s.expenseRecordService.GetExpenseRecords(ctx)
	if err != nil {
		return fmt.Errorf("error fetching expense records: %w", err)
	}

	s.expenseRecords = records

	return nil
}

func (s *DashboardService) getBankAccountBalance(ctx context.Context, userID *string) {

	balances := make(map[string]float64)
	for _, income := range s.incomeRecords {
		balances[income.BankAccountID] += income.Amount
	}
	for _, expense := range s.expenseRecords {
		if expense.DueDate.Before(utils.GetFirstDayOfLastMonth()) {
			balances[expense.BankPaidFrom] -= expense.Amount
		}
	}

	count := 0
	for bankID := range balances {
		fmt.Println("\n Count is:", count)
		count++

		if bankID == "" {
			continue
		}
		fmt.Println("\n Bank account ID:", bankID)

		banks, err := s.bankAccountService.GetBankAccounts(ctx)
		fmt.Println("\n Bank account name:", banks, "error:", err)
		if err != nil {
			continue
		}

		// fmt.Println("\n Bank account name:", name.ID, name.Description, name.BankCode)
		// if name == nil {
		// 	continue
		// }

		// fmt.Println("\n Bank account name:", name.ID, name.Description, name.BankCode)
		// s.dash.SummaryCards.AccountBalances = append(s.dash.SummaryCards.AccountBalances, dashboardEntity.AccountBalanceItem{
		// 	AccountName: name.Description,
		// 	BankName:    name.Description,
		// 	Balance:     balances[bankID],
		// 	UserID:      *userID,
		// })
	}

	fmt.Println("\n Count bank account balances:", len(s.dash.SummaryCards.AccountBalances))
	fmt.Println("\n Bank account balances response:", s.dash.SummaryCards.AccountBalances)
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

	amount := 0.0
	for _, income := range allUserIncomes {
		amount += income.Amount
	}
	log.Println("\n total income amount:", amount)

	allUserRawExpenses, err := s.expenseRecordService.GetExpenseRecords(ctx)
	if err != nil {
		fmt.Printf("Warning: Error fetching expense records for user %s: %v\n", userID, err)
		allUserRawExpenses = []financeEntity.ExpenseRecord{}
	}

	amount = 0.0
	for _, exp := range allUserRawExpenses {
		amount += exp.Amount
	}
	log.Println("\n total expense amount:", amount)

	allUserPaidExpenses := make([]financeEntity.ExpenseRecord, 0)
	for _, exp := range allUserRawExpenses {
		if !exp.PaymentDate.IsZero() {
			if !exp.PaymentDate.After(now) {
				allUserPaidExpenses = append(allUserPaidExpenses, exp)
			}
		}
	}

	// totalBalance := s.calculateTotalBalance(allUserBankAccounts, allUserIncomes, allUserPaidExpenses)
	monthlyRevenue := s.calculateMonthlyRevenue(allUserIncomes, currentMonthStart)
	monthlyExpenses := s.calculateMonthlyExpenses(allUserPaidExpenses, currentMonthStart)

	goalsProgressStr := "N/A (Data unavailable)"
	// profileGoals, err := s.profileGoalsService.GetProfileGoals(ctx, &userID)

	// if err != nil {
	// 	fmt.Printf("Warning: Error fetching profile goals for user %s: %v\n", userID, err)
	// } else {
	// 	goalsProgressStr = s.formatGoalsProgress(profileGoals)
	// }

	dashboard := &dashboardEntity.Dashboard{
		SummaryCards: dashboardEntity.SummaryCards{
			// TotalBalance:    totalBalance,
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
// func (s *DashboardService) calculateTotalBalance(
//
//	accounts []financeEntity.BankAccountRequest,
//	incomes []financeEntity.IncomeRecord,
//	paidExpenses []financeEntity.ExpenseRecord,
//
// ) float64 {
func (s *DashboardService) calculateTotalBalance(ctx context.Context, userID string) {
	accounts, err := s.bankAccountService.GetBankAccounts(ctx)
	if err != nil {
		fmt.Printf("Warning: Error fetching bank accounts for user %s: %v\n", userID, err)
		accounts = []financeEntity.BankAccountRequest{}
	}

	var totalBalance float64
	accountBalances := s.calculateAllAccountBalances(accounts, s.incomeRecords, s.expenseRecords)
	for _, balance := range accountBalances {
		totalBalance += balance
	}

	s.dash.SummaryCards.TotalBalance = totalBalance

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
		if expense.BankPaidFrom != "" {
			accountBalances[expense.BankPaidFrom] -= expense.Amount
		}
	}
	return accountBalances
}

func (s *DashboardService) calculateMonthlyRevenue(incomes []financeEntity.IncomeRecord, monthStart time.Time) float64 {
	var totalRevenue float64
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	for _, income := range incomes {
		if !income.ReceiptDate.Before(monthStart) && !income.ReceiptDate.After(monthEnd) {
			totalRevenue += income.Amount
		}
	}
	return totalRevenue
}

func (s *DashboardService) calculateMonthlyExpenses(paidExpenses []financeEntity.ExpenseRecord, monthStart time.Time) float64 {
	var totalExpenses float64
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	for _, expense := range paidExpenses {
		if !expense.PaymentDate.IsZero() {
			if !expense.PaymentDate.Before(monthStart) && !expense.PaymentDate.After(monthEnd) {
				totalExpenses += expense.Amount
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

func (s *DashboardService) formatGoalsProgress(ctx context.Context, userID string) {
	s.dash.SummaryCards.GoalsProgress = "N/A (Data unavailable)"
	profileGoals, err := s.profileGoalsService.GetProfileGoals(ctx, &userID)
	if err != nil {
		fmt.Printf("Warning: Error fetching profile goals for user %s: %v\n", userID, err)
	}

	allGoals := append(profileGoals.Goals2Years, profileGoals.Goals5Years...)
	allGoals = append(allGoals, profileGoals.Goals10Years...) // Typo: allGolas -> allGoals
	totalGoals := len(allGoals)
	completedGoals := 0 // Limitation: Cannot determine completed goals
	if totalGoals == 0 {
		s.dash.SummaryCards.GoalsProgress = "Nenhuma meta definida"
	}
	percentage := 0.0
	s.dash.SummaryCards.GoalsProgress = fmt.Sprintf("%.0f%% (%d de %d metas)", percentage, completedGoals, totalGoals)
}

func (s *DashboardService) getUpcomingBills2() error {

	bills := make([]dashboardEntity.UpcomingBill, 0)
	for _, expense := range s.expenseRecords {
		if expense.PaymentDate.IsZero() {
			bills = append(bills, dashboardEntity.UpcomingBill{
				BillName: fmt.Sprintf("%s - %s", expense.Category, expense.Subcategory),
				Amount:   expense.Amount,
				DueDate:  expense.DueDate.Format("2006-01-02"),
			})

		}
	}

	s.dash.UpcomingBillsData = bills
	sort.Slice(s.dash.UpcomingBillsData, func(i, j int) bool {
		// Parse DueDate strings back to time.Time for comparison
		dateI, _ := time.Parse("2006-01-02", s.dash.UpcomingBillsData[i].DueDate)
		dateJ, _ := time.Parse("2006-01-02", s.dash.UpcomingBillsData[j].DueDate)
		return dateI.Before(dateJ)
	})

	return nil
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
		if exp.PaymentDate.IsZero() {
			if !exp.DueDate.Before(upcomingStartDate) && !exp.DueDate.After(upcomingEndDate) {
				if exp.PaymentDate.IsZero() {
					billName := exp.Description
					if billName == "" {
						billName = exp.Category
					}
					bills = append(bills, dashboardEntity.UpcomingBill{
						BillName: billName,
						Amount:   exp.Amount,
						DueDate:  exp.DueDate.Format("2006-01-02"), // Assign the parsed time.Time value
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
			Month:    monthStart.Format("2006-01"),
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
		if !exp.PaymentDate.IsZero() {

			if !exp.PaymentDate.Before(monthStart) && !exp.PaymentDate.After(monthEnd) {
				categoryName := exp.Category
				if categoryName == "" {
					categoryName = "Outros"
				}
				categories[categoryName] += exp.Amount

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

func (s *DashboardService) accountBalance(ctx context.Context) {

	s.messageQueue.Consumer(ctx, mq_exchange, mq_queue_income, s.processIncomeRecord)
}

func (s *DashboardService) processIncomeRecord(body []byte, traceID string) error {
	var incomeRecord financeEntity.IncomeRecordEvent
	if err := json.Unmarshal(body, &incomeRecord); err != nil {
		return fmt.Errorf("erro ao deserializar: %w", err)
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "UserID", incomeRecord.Data.UserID)

	// Processar baseado no contexto ou adicionar campo Action na struct
	var balance float64 = 0.0

	switch {
	case incomeRecord.Action == entity_common.ActionCreate:
		balance += incomeRecord.Data.Amount
	case incomeRecord.Action == entity_common.ActionDelete:
		balance += (-incomeRecord.Data.Amount)
	default:
		return errors.New("action did not match any case")
	}

	platfotmInst, err := s.platformInstitution.GetAllFinancialInstitutions(ctx)
	if err != nil {
		return err
	}

	if len(platfotmInst) == 0 {
		return errors.New("financial institution not found")
	}

	bankAccount, err := s.bankAccountService.GetByFilter(ctx, map[string]interface{}{"id": incomeRecord.Data.BankAccountID})
	if err != nil {
		return err
	}

	var bankName string
	if len(bankAccount) > 0 && len(platfotmInst) > 0 {
		for _, v := range platfotmInst {
			if v.Code == bankAccount[0].BankCode {
				bankName = v.Name
				break
			}
		}
	}

	dashboard, err := s.dashboardRepository.GetBankAccountBalanceByID(ctx, &incomeRecord.Data.UserID, &bankName)
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return err
		}
	}

	if dashboard == nil {
		dashboard = &dashboardEntity.AccountBalanceItem{
			UserID:      incomeRecord.Data.UserID,
			AccountName: bankAccount[0].Description,
			BankName:    bankName,
			Balance:     0.0,
		}
	}
	dashboard.Balance += balance
	s.dashboardRepository.UpdateBankAccountBalance(ctx, &incomeRecord.Data.UserID, dashboard)

	return nil
}

func (s *DashboardService) getMonthlyFinancialSummary(ctx context.Context, userID *string) ([]dashboardEntity.MonthlyFinancialSummaryItem, error) {
	if userID == nil || *userID == "" {
		return nil, fmt.Errorf("userID is nil or empty")
	}

	now := time.Now()
	// Calculate the start date 12 months ago from the beginning of the current month
	endDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, 1, 0).Add(-time.Nanosecond) // End of current month
	startDate := endDate.AddDate(0, -12, 1)                                                                             // Start of the month 12 months ago

	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	// Fetch all relevant income and expense records in one go
	allUserIncomes, err := s.incomeRecordService.GetIncomeRecords(ctx, &financeEntity.GetIncomeRecordsQueryParameters{
		StartDate: &startDateStr,
		EndDate:   &endDateStr,
		UserID:    *userID,
	})
	if err != nil {
		// Log the error and continue, perhaps with partial data or empty result
		fmt.Printf("Warning: Error fetching income records for financial summary for user %s: %v\n", *userID, err)
		allUserIncomes = []financeEntity.IncomeRecord{} // Ensure it's not nil
	}

	// Assuming ExpenseRecordService also has a method to get records by date range and UserID
	// If not, you might need to fetch all and filter by UserID manually
	allUserExpenses, err := s.expenseRecordService.GetExpenseRecordsByDate(ctx, &financeEntity.ExpenseRecordQueryByDate{
		StartDate: startDateStr,
		EndDate:   endDateStr,
	})
	if err != nil {
		// Log the error and continue
		fmt.Printf("Warning: Error fetching expense records for financial summary for user %s: %v\n", *userID, err)
		allUserExpenses = []financeEntity.ExpenseRecord{} // Ensure it's not nil
	}

	// Aggregate data by month
	monthlySummaryMap := make(map[string]*dashboardEntity.MonthlyFinancialSummaryItem)

	// Process incomes
	for _, income := range allUserIncomes {

		// Ensure date is within the desired range (should be covered by the initial fetch, but good practice)
		if income.ReceiptDate.Before(startDate) || income.ReceiptDate.After(endDate) {
			continue
		}

		monthLabel := income.ReceiptDate.Format("2006-01")
		if item, exists := monthlySummaryMap[monthLabel]; exists {
			item.TotalIncome += income.Amount
		} else {
			monthlySummaryMap[monthLabel] = &dashboardEntity.MonthlyFinancialSummaryItem{
				Month:       monthLabel,
				TotalIncome: income.Amount,
			}
		}
	}

	// Process expenses
	for _, expense := range allUserExpenses {
		// Only consider paid expenses for the monthly summary
		if !expense.PaymentDate.IsZero() {

			// Ensure date is within the desired range
			if expense.PaymentDate.Before(startDate) || expense.PaymentDate.After(endDate) {
				continue
			}

			monthLabel := expense.PaymentDate.Format("2006-01")
			if item, exists := monthlySummaryMap[monthLabel]; exists {
				item.TotalExpenses += expense.Amount
			} else {
				// This case might happen if there are expenses in a month with no income
				monthlySummaryMap[monthLabel] = &dashboardEntity.MonthlyFinancialSummaryItem{
					Month:         monthLabel,
					TotalExpenses: expense.Amount,
				}
			}
		}
	}

	// Convert map to slice and sort by month
	monthlySummary := make([]dashboardEntity.MonthlyFinancialSummaryItem, 0, len(monthlySummaryMap))
	for _, item := range monthlySummaryMap {
		monthlySummary = append(monthlySummary, *item)
	}

	// Sort the summary by month (chronologically)
	sort.Slice(monthlySummary, func(i, j int) bool {
		// Parse month labels back to time.Time for accurate sorting
		dateI, errI := time.Parse("2006-01", monthlySummary[i].Month)
		dateJ, errJ := time.Parse("2006-01", monthlySummary[j].Month)

		// Handle parsing errors - if error, consider that item "later" in the sort
		if errI != nil && errJ != nil {
			return false
		} // Both invalid, order doesn't matter for sorting
		if errI != nil {
			return false
		} // i is invalid, j is valid, j comes first
		if errJ != nil {
			return true
		} // j is invalid, i is valid, i comes first

		return dateI.Before(dateJ)
	})

	// Optional: Update the repository with the calculated summary
	// This part depends on whether you need to persist this aggregated summary.
	// If so, you can iterate through the 'monthlySummary' slice and call a repository update method.
	// For simplicity and avoiding potential race conditions, do this sequentially here
	// instead of in separate goroutines as before.
	for _, summaryItem := range monthlySummary {
		// Implement a repository method to update/create the monthly summary record
		s.updateMonthlyFinancialSummary(ctx, userID, &summaryItem)
	}

	return monthlySummary, nil
}

func (s *DashboardService) monthlyFinancialSummaryIncome(ctx context.Context, startDate, endDate, userID *string) []financeEntity.IncomeRecord {

	allUserIncomes, _ := s.incomeRecordService.GetIncomeRecords(ctx, &financeEntity.GetIncomeRecordsQueryParameters{
		StartDate: startDate,
		EndDate:   endDate,
		UserID:    *userID,
	})

	return allUserIncomes
}

func (s *DashboardService) monthlyFinancialSummaryExpense(ctx context.Context, startDate, endDate, userID *string) []financeEntity.ExpenseRecord {
	allUserExpenses, _ := s.expenseRecordService.GetExpenseRecordsByDate(ctx, &financeEntity.ExpenseRecordQueryByDate{
		StartDate: *startDate,
		EndDate:   *endDate,
	})

	return allUserExpenses
}

func (s *DashboardService) updateMonthlyFinancialSummary(ctx context.Context, userID *string, data *dashboardEntity.MonthlyFinancialSummaryItem) {

	if userID == nil {
		return
	}

	result, err := s.dashboardRepository.GetFinancialSummary(ctx, userID)
	if err != nil {
		return
	}

	needUpdate := false
	exists := false
	for _, v := range result {

		if v.Month == data.Month {
			exists = true
			data.ID = v.ID
			if v.TotalIncome != data.TotalIncome {
				v.TotalIncome = data.TotalIncome
				needUpdate = true
			}
			if v.TotalExpenses != data.TotalExpenses {
				v.TotalExpenses = data.TotalExpenses
				needUpdate = true
			}
			break
		}
	}

	if needUpdate || !exists {
		if !exists {
			data.CreatedAt = time.Now()
		}
		data.UpdatedAt = time.Now()
		data.UserID = *userID
		s.dashboardRepository.UpdateFinancialSummary(ctx, userID, data)
	}

}
