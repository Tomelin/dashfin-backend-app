// Package dashboard defines the repository for aggregating dashboard data.
package dashboard

import (
	"context"
	"time"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
	repo_finance "github.com/Tomelin/dashfin-backend-app/internal/core/repository/finance"
	repo_platform "github.com/Tomelin/dashfin-backend-app/internal/core/repository/platform"
	repo_profile "github.com/Tomelin/dashfin-backend-app/internal/core/repository/profile"
)

const (
	defaultUpcomingBillsLimit       = 5
	defaultRecommendationsLimit     = 3
	defaultRevenueExpenseChartPeriods = 6 // e.g., 6 months
)

// RepositoryInterface defines methods for fetching aggregated dashboard data.
type RepositoryInterface interface {
	GetSummaryCardsData(ctx context.Context, userID string) (*entity_dashboard.SummaryCards, error)
	GetAccountSummaries(ctx context.Context, userID string) ([]entity_dashboard.AccountSummaryData, error)
	GetUpcomingBills(ctx context.Context, userID string) ([]entity_dashboard.UpcomingBill, error)
	GetRevenueExpenseChartData(ctx context.Context, userID string) ([]entity_dashboard.RevenueExpenseChartData, error)
	GetExpenseCategoryChartData(ctx context.Context, userID string) ([]entity_dashboard.ExpenseCategoryChartData, error)
	GetPersonalizedRecommendations(ctx context.Context, userID string) ([]entity_dashboard.PersonalizedRecommendation, error)
}

// dashboardRepository implements RepositoryInterface and holds references to other repositories.
type dashboardRepository struct {
	accountRepo      repo_finance.AccountRepositoryInterface
	transactionRepo  repo_finance.TransactionRepositoryInterface
	goalRepo         repo_profile.GoalRepositoryInterface
	recommendationRepo repo_platform.RecommendationRepositoryInterface
}

// NewDashboardRepository creates a new instance of dashboardRepository.
func NewDashboardRepository(
	accountRepo repo_finance.AccountRepositoryInterface,
	transactionRepo repo_finance.TransactionRepositoryInterface,
	goalRepo repo_profile.GoalRepositoryInterface,
	recommendationRepo repo_platform.RecommendationRepositoryInterface,
) RepositoryInterface {
	return &dashboardRepository{
		accountRepo:      accountRepo,
		transactionRepo:  transactionRepo,
		goalRepo:         goalRepo,
		recommendationRepo: recommendationRepo,
	}
}

// GetSummaryCardsData fetches and aggregates data for the summary cards.
func (r *dashboardRepository) GetSummaryCardsData(ctx context.Context, userID string) (*entity_dashboard.SummaryCards, error) {
	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	totalBalance, err := r.accountRepo.GetTotalBalanceByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	totalIncome, err := r.transactionRepo.GetMonthlyRevenueByUserID(ctx, userID, currentYear, currentMonth)
	if err != nil {
		return nil, err
	}

	totalExpenses, err := r.transactionRepo.GetMonthlyExpensesByUserID(ctx, userID, currentYear, currentMonth)
	if err != nil {
		return nil, err
	}

	// Placeholder for SavingsGoal - Assuming it might come from goalRepo or be a fixed value.
	// For now, let's use a placeholder. This should be refined based on actual logic for savings goal.
	// goalsProgress, err := r.goalRepo.GetGoalsProgressByUserID(ctx, userID)
	// if err != nil {
	// 	return nil, err
	// }
	// For now, using a dummy value for SavingsGoal as its source isn't fully defined by GetGoalsProgressByUserID (string)
	savingsGoal := 1000.00 // Placeholder

	upcomingBillsList, err := r.transactionRepo.GetUpcomingBillsByUserID(ctx, userID, defaultUpcomingBillsLimit)
	if err != nil {
		return nil, err
	}
	upcomingBillsCount := len(upcomingBillsList)

	// Placeholder for UncategorizedTx - This would typically come from the transaction repository
	// with a method like GetUncategorizedTransactionsCountByUserID.
	uncategorizedTxCount := 0 // Placeholder

	return &entity_dashboard.SummaryCards{
		TotalBalance:    totalBalance,
		TotalIncome:     totalIncome,
		TotalExpenses:   totalExpenses,
		SavingsGoal:     savingsGoal, // Replace with actual logic
		UpcomingBills:   upcomingBillsCount,
		UncategorizedTx: uncategorizedTxCount, // Replace with actual logic
	}, nil
}

// GetAccountSummaries fetches account summaries.
func (r *dashboardRepository) GetAccountSummaries(ctx context.Context, userID string) ([]entity_dashboard.AccountSummaryData, error) {
	return r.accountRepo.GetAccountSummariesByUserID(ctx, userID)
}

// GetUpcomingBills fetches upcoming bills.
func (r *dashboardRepository) GetUpcomingBills(ctx context.Context, userID string) ([]entity_dashboard.UpcomingBill, error) {
	return r.transactionRepo.GetUpcomingBillsByUserID(ctx, userID, defaultUpcomingBillsLimit)
}

// GetRevenueExpenseChartData fetches data for the revenue vs. expense chart.
func (r *dashboardRepository) GetRevenueExpenseChartData(ctx context.Context, userID string) ([]entity_dashboard.RevenueExpenseChartData, error) {
	return r.transactionRepo.GetRevenueExpenseChartDataByUserID(ctx, userID, defaultRevenueExpenseChartPeriods)
}

// GetExpenseCategoryChartData fetches data for the expense by category chart.
func (r *dashboardRepository) GetExpenseCategoryChartData(ctx context.Context, userID string) ([]entity_dashboard.ExpenseCategoryChartData, error) {
	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())
	return r.transactionRepo.GetExpenseCategoryChartDataByUserID(ctx, userID, currentYear, currentMonth)
}

// GetPersonalizedRecommendations fetches personalized recommendations.
func (r *dashboardRepository) GetPersonalizedRecommendations(ctx context.Context, userID string) ([]entity_dashboard.PersonalizedRecommendation, error) {
	return r.recommendationRepo.GetPersonalizedRecommendationsByUserID(ctx, userID, defaultRecommendationsLimit)
}
