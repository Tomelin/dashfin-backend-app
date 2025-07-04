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

	s.getMonthlyFinancialSummary(&userID)

	return &s.dash, nil
}

func (s *DashboardService) getSummaryCards() error {

	var receiveMonth float64
	var expenseMonth float64
	var receiveBalance float64
	var expenseBalance float64
	var receiveLastMonth float64
	var expenseLastMonth float64

	for _, income := range s.incomeRecords {
		if income.ReceiptDate.After(utils.GetFirstDayOfCurrentMonth()) && income.ReceiptDate.Before(utils.GetLastDayOfCurrentMonth()) {
			receiveMonth += income.Amount
		}
		receiveBalance += income.Amount
	}

	for _, expense := range s.expenseRecords {
		if expense.DueDate.After(utils.GetFirstDayOfCurrentMonth()) && expense.DueDate.Before(utils.GetLastDayOfCurrentMonth()) {
			expenseMonth += expense.Amount
		}
		if expense.DueDate.Before(utils.GetLastDayOfCurrentMonth()) {
			expenseBalance += expense.Amount
		}
	}

	for _, income := range s.incomeRecords {
		if income.ReceiptDate.After(utils.GetFirstDayOfLastMonth()) && income.ReceiptDate.Before(utils.GetLastDayOfLastMonth()) {
			receiveLastMonth += income.Amount
		}
	}

	for _, expense := range s.expenseRecords {
		if expense.DueDate.After(utils.GetFirstDayOfLastMonth()) && expense.DueDate.Before(utils.GetLastDayOfLastMonth()) {
			expenseLastMonth += expense.Amount
		}
	}

	totalBalance := receiveBalance - expenseBalance
	totalBalanceLastMonth := receiveLastMonth - expenseLastMonth

	s.dash.SummaryCards.MonthlyExpensesChangePercent = ((expenseMonth / expenseLastMonth) - 1) * 100
	s.dash.SummaryCards.MonthlyRevenueChangePercent = ((receiveMonth / receiveLastMonth) - 1) * 100
	s.dash.SummaryCards.TotalBalanceChangePercent = ((totalBalance / totalBalanceLastMonth) - 1) * 100
	s.dash.SummaryCards.MonthlyExpenses = expenseMonth
	s.dash.SummaryCards.MonthlyRevenue = receiveMonth
	s.dash.SummaryCards.TotalBalance = totalBalance

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

func (s *DashboardService) getIncomeRecordsFromPeriod(startDate, endDate time.Time) ([]financeEntity.IncomeRecord, float64, error) {

	if s.incomeRecords == nil {
		return nil, 0, errors.New("income records are not initialized")
	}

	if startDate.IsZero() || endDate.IsZero() {
		return nil, 0, fmt.Errorf("startDate and endDate must be provided")
	}

	var amount float64
	var records []financeEntity.IncomeRecord
	for _, income := range s.incomeRecords {
		if income.ReceiptDate.After(startDate) && income.ReceiptDate.Before(endDate) {
			records = append(records, income)
			amount += income.Amount
		}
	}

	return records, amount, nil
}

func (s *DashboardService) getExpenseRecords(ctx context.Context) error {

	records, err := s.expenseRecordService.GetExpenseRecords(ctx)
	if err != nil {
		return fmt.Errorf("error fetching expense records: %w", err)
	}

	s.expenseRecords = records

	return nil
}

func (s *DashboardService) getExpenseRecordsFromPeriod(startDate, endDate time.Time) ([]financeEntity.ExpenseRecord, float64, error) {

	if s.expenseRecords == nil {
		return nil, 0, errors.New("expense records are not initialized")
	}

	if startDate.IsZero() || endDate.IsZero() {
		return nil, 0, fmt.Errorf("startDate and endDate must be provided")
	}

	var amount float64
	var records []financeEntity.ExpenseRecord
	for _, expense := range s.expenseRecords {
		if expense.DueDate.After(startDate) && expense.DueDate.Before(endDate) {
			records = append(records, expense)
			amount += expense.Amount
		}
	}

	return records, amount, nil
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

	banks, err := s.bankAccountService.GetBankAccounts(ctx)
	if err != nil {
		return
	}

	for bankID := range balances {

		if balances[bankID] == 0 {
			continue
		}

		for _, bank := range banks {

			if bank.ID == bankID {
				s.dash.SummaryCards.AccountBalances = append(s.dash.SummaryCards.AccountBalances, dashboardEntity.AccountBalanceItem{
					AccountName: bank.CustomBankName,
					BankName:    bank.Description,
					Balance:     balances[bankID],
					UserID:      *userID,
					ID:          bank.ID,
				})
			}
		}

	}
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

func (s *DashboardService) getMonthlyFinancialSummary(userID *string) error {

	if userID == nil || *userID == "" {
		return fmt.Errorf("userID is nil or empty")
	}

	items := make([]dashboardEntity.MonthlyFinancialSummaryItem, 0)

	for i := 0; i < 12; i++ {
		startDateThisMonth := utils.GetFirstDayOfCurrentMonth().AddDate(0, -i, 0)
		endDateThisMonth := utils.GetLastDayOfCurrentMonth().AddDate(0, -i, 0)
		month := startDateThisMonth.Format("2006-01")

		_, incomeAmount, err := s.getIncomeRecordsFromPeriod(startDateThisMonth, endDateThisMonth)
		if err != nil {
			return fmt.Errorf("error fetching income records for month %s: %w", month, err)
		}

		_, expenseAmount, err := s.getExpenseRecordsFromPeriod(startDateThisMonth, endDateThisMonth)
		if err != nil {
			return fmt.Errorf("error fetching expense records for month %s: %w", month, err)
		}

		items = append(items, dashboardEntity.MonthlyFinancialSummaryItem{
			Month:         month,
			TotalIncome:   incomeAmount,
			TotalExpenses: expenseAmount,
			UserID:        *userID,
		})
	}

	s.dash.SummaryCards.MonthlyFinancialSummary = items
	return nil

}
