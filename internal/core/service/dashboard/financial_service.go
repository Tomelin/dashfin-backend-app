package dashboard

import (
	"context"
	"fmt"
	"log"
	"math"
	"regexp"
	"sort"

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
	repo    repo_dashboard.FinancialRepositoryInterface // Interface for accessing financial data.
	expsene finance_entity.ExpenseRecordServiceInterface
	cache   cache.CacheService
}

// NewFinancialService creates and returns a new instance of FinancialService.
// It requires a FinancialRepositoryInterface for data access and a FirebaseDBInterface
// to get the database client.
func NewFinancialService(repo repo_dashboard.FinancialRepositoryInterface, cache cache.CacheService, expense finance_entity.ExpenseRecordServiceInterface) (FinancialServiceInterface, error) {

	if repo == nil {
		return nil, fmt.Errorf("repo cannot be nil")
	}

	if cache == nil {
		return nil, fmt.Errorf("cache cannot be nil")
	}

	return &FinancialService{
		repo:    repo,
		cache:   cache,
		expsene: expense,
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
	log.Printf("GetPlannedVsActual called for userID: %s, Month: %d, Year: %d", userID, req.Month, req.Year)

	// Determine Month and Year
	currentMonth := req.Month
	currentYear := req.Year
	if currentMonth == 0 {
		currentMonth = int(time.Now().Month())
		log.Printf("Defaulting Month to current month: %d", currentMonth)
	}
	if currentYear == 0 {
		currentYear = time.Now().Year()
		log.Printf("Defaulting Year to current year: %d", currentYear)
	}

	// Fetch Data
	log.Printf("Fetching expense planning for userID: %s, Month: %d, Year: %d", userID, currentMonth, currentYear)

	recors, err := s.expsene.GetExpenseRecords(ctx)
	if err != nil {
		log.Printf("Error fetching expense records: %v", err)
		return nil, fmt.Errorf("failed to get expense records: %w", err)
	}
	log.Printf("Loaded %d expense records", len(recors))
	for _, v := range recors {
		b, err := isInCurrentMonthAndYear(v.DueDate)
		log.Println(v.DueDate, b, err)
	}

	planningDoc, err := s.repo.GetExpensePlanning(ctx, userID, currentMonth, currentYear)
	if err != nil {
		log.Printf("Error fetching expense planning for userID %s: %v", userID, err)
		return nil, fmt.Errorf("failed to get expense planning: %w", err)
	}
	if planningDoc == nil {
		log.Printf("No expense planning found for userID %s for %d/%d. Returning empty result.", userID, currentMonth, currentYear)
		return []entity_dashboard.PlannedVsActualCategory{}, nil // No planning, return empty as per 404 requirement
	}
	if planningDoc.Categories == nil || len(planningDoc.Categories) == 0 {
		log.Printf("Expense planning found for userID %s for %d/%d, but no categories defined. Returning empty result.", userID, currentMonth, currentYear)
		return []entity_dashboard.PlannedVsActualCategory{}, nil
	}

	log.Printf("Fetching actual expenses for userID: %s, Month: %d, Year: %d", userID, currentMonth, currentYear)
	actualExpenseDocs, err := s.repo.GetExpenses(ctx, userID, currentMonth, currentYear)
	if err != nil {
		log.Printf("Error fetching actual expenses for userID %s: %v", userID, err)
		return nil, fmt.Errorf("failed to get actual expenses: %w", err)
	}

	log.Printf("Fetching expense categories definitions")
	categoryDocs, err := s.repo.GetExpenseCategories(ctx)
	if err != nil {
		log.Printf("Error fetching expense categories definitions: %v", err)
		return nil, fmt.Errorf("failed to get expense categories: %w", err)
	}

	categoryLabels := make(map[string]string)
	for _, catDoc := range categoryDocs {
		categoryLabels[catDoc.Category] = catDoc.Label
	}
	log.Printf("Loaded %d category labels", len(categoryLabels))

	// Process Data
	actualAmounts := make(map[string]float64)
	for _, expense := range actualExpenseDocs {
		if !categoryRegex.MatchString(expense.Category) {
			log.Printf("Warning: Actual expense document ID %s has invalid category key format: %s. Skipping.", expense.ID, expense.Category)
			continue // Skip if category key in actual expense is invalid
		}
		actualAmounts[expense.Category] += expense.Amount
	}
	log.Printf("Aggregated %d actual expense amounts by category", len(actualAmounts))

	var result []entity_dashboard.PlannedVsActualCategory

	log.Println("Processing planned categories...")
	for categoryKey, plannedAmount := range planningDoc.Categories {
		// Validation: category key format
		if !categoryRegex.MatchString(categoryKey) {
			log.Printf("Warning: Planned category key '%s' does not match format ^[a-z0-9_]+$. Skipping this category.", categoryKey)
			continue
		}

		// Validation: plannedAmount must be non-negative (already ensured by struct tag gte=0 if source is validated)
		if plannedAmount < 0 {
			log.Printf("Warning: Planned amount for category '%s' is negative (%.2f). Skipping this category.", categoryKey, plannedAmount)
			continue
		}

		actualAmount := actualAmounts[categoryKey] // Defaults to 0.0 if not present

		label, labelExists := categoryLabels[categoryKey]
		if !labelExists || label == "" {
			log.Printf("Warning: No label found for category key '%s'. Defaulting label to category key.", categoryKey)
			label = categoryKey // Default label to categoryKey if not found or empty
		}

		var spentPercentage float64
		if plannedAmount == 0 {
			spentPercentage = 0 // If nothing was planned, spent percentage is 0, even if there was actual spending.
		} else {
			spentPercentage = (actualAmount / plannedAmount) * 100
		}
		spentPercentage = roundToTwoDecimals(spentPercentage)

		// Validation: actualAmount must be non-negative (already ensured by struct tag gte=0 if source is validated)
		// No explicit check here as it's summed from validated or trusted sources.

		pvaCategory := entity_dashboard.PlannedVsActualCategory{
			Category:        categoryKey,
			Label:           label,
			PlannedAmount:   roundToTwoDecimals(plannedAmount), // Ensure planned amount is also rounded for consistency
			ActualAmount:    roundToTwoDecimals(actualAmount),
			SpentPercentage: spentPercentage,
		}
		result = append(result, pvaCategory)
	}

	// Sort Results by Category key for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Category < result[j].Category
	})
	log.Printf("Processed %d categories for PlannedVsActual response.", len(result))

	return result, nil
}

func isInCurrentMonthAndYear(dateString string) (bool, error) {
	// 1. Parse da string para time.Time
	// Usamos o layout "2006-01-02", que é a forma padrão do Go para especificar formatos de data.
	// Isso garante que a string seja interpretada corretamente.
	inputDate, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		// Retornamos um erro claro se o formato for inválido, tornando a função mais robusta.
		return false, fmt.Errorf("erro ao analisar a data: %w", err)
	}

	// 2. Obter a data e hora atuais
	now := time.Now()

	// 3. Comparar ano e mês
	// Acessamos e comparamos o ano e o mês de ambas as datas.
	// Esta é a forma mais clara e performática de fazer essa verificação específica.
	isSameYear := inputDate.Year() == now.Year()
	isSameMonth := inputDate.Month() == now.Month()

	return isSameYear && isSameMonth, nil
}
