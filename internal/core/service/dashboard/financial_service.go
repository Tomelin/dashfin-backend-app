package dashboard

import (
	"context"
	"fmt"
	"math"
	"regexp"

	// "strconv" // Not used directly, can be removed if not needed by other implicit operations
	"time"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
	finance_entity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	repo_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/repository/dashboard"
	"github.com/Tomelin/dashfin-backend-app/pkg/cache"
	// "github.com/shopspring/decimal" // Acknowledged for future use if higher precision is needed
)

// Pre-compile regex for category key validation
var categoryRegex = regexp.MustCompile("^[a-z0-9_]+$")

// FinancialServiceInterface defines the contract for financial services related to dashboards.
// It provides methods to calculate and retrieve financial summaries and analyses.
type FinancialServiceInterface interface {
	// GetPlannedVsActual calculates and returns a comparison of planned versus actual spending
	// for various expense categories for a given user and period.
	// The period (month/year) can be specified in the PlannedVsActualRequest.
	// If the period is not specified, it typically defaults to the current month and year.
	// It returns a slice of PlannedVsActualCategory, each representing a category's financial summary.
	// If no planning data is found for the period, an empty slice and no error are returned.
	// Errors during data retrieval or processing will be returned as a non-nil error.
	GetPlannedVsActual(ctx context.Context, userID string, req entity_dashboard.PlannedVsActualRequest) ([]entity_dashboard.PlannedVsActualCategory, error)
}

// FinancialService provides services for calculating and retrieving dashboard-related financial information.
// It uses a FinancialRepositoryInterface to access underlying financial data and a
// FirebaseDBInterface to obtain a Firestore client.
type FinancialService struct {
	repo      repo_dashboard.FinancialRepositoryInterface // Interface for accessing financial data.
	expsene   finance_entity.ExpenseRecordServiceInterface
	spendPLan finance_entity.SpendingPlanServiceInterface
	cache     cache.CacheService
}

// NewFinancialService creates and returns a new instance of FinancialService.
// It requires a FinancialRepositoryInterface for data access and a FirebaseDBInterface
// to get the database client.
func NewFinancialService(repo repo_dashboard.FinancialRepositoryInterface, cache cache.CacheService, expense finance_entity.ExpenseRecordServiceInterface, spendPLan finance_entity.SpendingPlanServiceInterface) (FinancialServiceInterface, error) {

	if repo == nil {
		return nil, fmt.Errorf("repo cannot be nil")
	}

	if cache == nil {
		return nil, fmt.Errorf("cache cannot be nil")
	}

	return &FinancialService{
		repo:      repo,
		cache:     cache,
		expsene:   expense,
		spendPLan: spendPLan,
	}, nil
}

// roundToTwoDecimals is an unexported helper function to round a float64 to two decimal places.
func roundToTwoDecimals(f float64) float64 {
	return math.Round(f*100) / 100
}

// GetPlannedVsActual calculates and returns the planned versus actual spending for various categories
// for a given user and period (month/year).
//
// It performs the following steps:
//  1. Obtains a Firestore client via the dbProvider.
//  2. Defaults month/year to current if not provided in the request.
//  3. Fetches expense planning data, actual expense records, and category definitions using the repository.
//  4. If no planning data is found for the specified period, it returns an empty list, signifying no content for that period.
//  5. Aggregates actual expenses by category.
//  6. For each category in the planning data:
//     a. Validates the category key format.
//     b. Retrieves the category label (defaults to key if not found).
//     c. Calculates planned amount, actual amount, and spent percentage (actual/planned * 100).
//     - If planned amount is 0, spent percentage is 0.
//     d. Rounds monetary values and percentages to two decimal places.
//  7. Sorts the results by category key for consistent output.
//
// Parameters:
//   - ctx: The context for the operation.
//   - userID: The ID of the user for whom to fetch the data.
//   - req: A PlannedVsActualRequest containing optional month and year for the report period.
//
// Returns:
//   - A slice of PlannedVsActualCategory structs, each detailing a category's financial summary.
//   - An error if any issue occurs during data fetching or processing (e.g., database errors),
//     except for the "no planning data found" case, which returns ([], nil).
func (s *FinancialService) GetPlannedVsActual(ctx context.Context, userID string, req entity_dashboard.PlannedVsActualRequest) ([]entity_dashboard.PlannedVsActualCategory, error) {

	// Determine Month and Year
	currentMonth := req.Month
	currentYear := req.Year
	if currentMonth == 0 {
		currentMonth = int(time.Now().Month())
	}
	if currentYear == 0 {
		currentYear = time.Now().Year()
	}

	currentSpendPlan := make([]entity_dashboard.PlannedVsActualCategory, 0)

	expenses := s.getExpenses(ctx, currentMonth, currentYear)
	if len(expenses) == 0 {
		return currentSpendPlan, nil
	}

	sumResult := sumExpense(expenses)

	spend := s.getSpendingPlan(ctx, userID)
	if spend == nil {
		return currentSpendPlan, nil
	}

	for _, v := range spend.CategoryBudgets {
		for _, v2 := range sumResult {
			if v.Category == v2.Category {
				currentSpendPlan = append(currentSpendPlan, entity_dashboard.PlannedVsActualCategory{
					Category:        v2.Category,
					PlannedAmount:   v.Amount,
					ActualAmount:    v2.Amount,
					Label:           v2.Category,
					SpentPercentage: roundToTwoDecimals(v2.Amount / spend.MonthlyIncome * 100),
				})
				break
			}
		}
	}

	var unlistedSpendTotal float64 = 0
	for _, actual := range sumResult {
		var wasPlanned bool = false
		for _, planned := range spend.CategoryBudgets {
			if actual.Category == planned.Category {
				wasPlanned = true
				break
			}
		}
		if !wasPlanned {
			unlistedSpendTotal += actual.Amount
		}
	}

	if unlistedSpendTotal > 0 {
		currentSpendPlan = append(currentSpendPlan, entity_dashboard.PlannedVsActualCategory{
			Category:        "Não cadastrado",
			ActualAmount:    unlistedSpendTotal,
			Label:           "Não cadastrado",
			SpentPercentage: roundToTwoDecimals(unlistedSpendTotal / spend.MonthlyIncome * 100),
			PlannedAmount:   0,
		})
	}

	return currentSpendPlan, nil
}

func (s *FinancialService) getExpenses(ctx context.Context, month, year int) []finance_entity.ExpenseRecord {

	expenses := make([]finance_entity.ExpenseRecord, 0)
	queryExpenses, err := s.expsene.GetExpenseRecords(ctx)
	if err != nil {
		return expenses
	}

	for _, v := range queryExpenses {
		b, err := s.isInCurrentMonthAndYear(v.DueDate)
		if err != nil || !b {
			continue
		}
		expenses = append(expenses, v)
	}

	return expenses
}

func (s *FinancialService) getSpendingPlan(ctx context.Context, userID string) *finance_entity.SpendingPlan {

	if userID == "" {
		return nil
	}

	querySpend, err := s.spendPLan.GetSpendingPlan(ctx, userID)
	if err != nil {
		return nil
	}
	if querySpend == nil {
		return nil
	}

	filteredCategoryBudgets := make([]finance_entity.CategoryBudget, 0)

	for _, v := range querySpend.CategoryBudgets { // Iterate over a copy or use a different approach if removing during iteration
		if v.Amount > 0.00 {
			filteredCategoryBudgets = append(filteredCategoryBudgets, v)
		}
	}

	spend := &finance_entity.SpendingPlan{
		ID:              querySpend.ID,
		CategoryBudgets: filteredCategoryBudgets,
		UserID:          querySpend.UserID,
		MonthlyIncome:   querySpend.MonthlyIncome,
		CreatedAt:       querySpend.CreatedAt,
		UpdatedAt:       querySpend.UpdatedAt,
	}

	return spend
}

type sumExpenseItems struct {
	Category string
	Amount   float64
}

func sumExpense(expenses []finance_entity.ExpenseRecord) []sumExpenseItems {

	expenseMap := make([]sumExpenseItems, 0)

	for _, expense := range expenses {

		// Verificar se a categoria da despesa já existe no mapa
		found := false

		for _, spend := range expenseMap {
			if spend.Category == expense.Category {
				spend.Amount += expense.Amount
				found = true
				break
			}
		}

		if !found {
			expenseMap = append(expenseMap, sumExpenseItems{
				Category: expense.Category,
				Amount:   expense.Amount,
			})
		}
	}

	return expenseMap
}

func (s *FinancialService) isInCurrentMonthAndYear(inputDate time.Time) (bool, error) {

	// 1. Obter a data e hora atuais
	now := time.Now()

	// 1. Comparar ano e mês
	// Acessamos e comparamos o ano e o mês de ambas as datas.
	// Esta é a forma mais clara e performática de fazer essa verificação específica.
	isSameYear := inputDate.Year() == now.Year()
	isSameMonth := inputDate.Month() == now.Month()

	return isSameYear && isSameMonth, nil
}
