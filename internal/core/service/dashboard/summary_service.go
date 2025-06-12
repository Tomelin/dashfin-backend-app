// Package dashboard defines the service for preparing the dashboard summary.
package dashboard

import (
	"context"
	"fmt"
	"time"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
	repo_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/repository/dashboard"
	"github.com/Tomelin/dashfin-backend-app/pkg/cache"
	// Import other necessary packages, like error handling utilities if you have them.
)

// ServiceInterface defines methods for the dashboard service.
type ServiceInterface interface {
	GetDashboardSummary(ctx context.Context, userID string) (*entity_dashboard.DashboardSummary, error)
}

// dashboardService implements ServiceInterface and uses a dashboard repository and a cache.
type dashboardService struct {
	dashboardRepo repo_dashboard.RepositoryInterface
	cache         cache.RedisCacheInterface
	cacheTTL      time.Duration
}

// NewDashboardService creates a new instance of dashboardService.
func NewDashboardService(
	dashboardRepo repo_dashboard.RepositoryInterface,
	redisCache cache.RedisCacheInterface,
	defaultTTLSeconds int,
) ServiceInterface {
	return &dashboardService{
		dashboardRepo: dashboardRepo,
		cache:         redisCache,
		cacheTTL:      time.Duration(defaultTTLSeconds) * time.Second,
	}
}

// GetDashboardSummary orchestrates calls to the dashboard repository to fetch all data,
// using caching to improve performance.
func (s *dashboardService) GetDashboardSummary(ctx context.Context, userID string) (*entity_dashboard.DashboardSummary, error) {
	cacheKey := fmt.Sprintf("dashboard_summary:%s", userID)
	var summary entity_dashboard.DashboardSummary

	// 1. Try to get data from cache
	if err := s.cache.Get(ctx, cacheKey, &summary); err == nil {
		// Cache hit
		return &summary, nil
	} else if err != cache.ErrCacheMiss {
		// Handle other cache errors (e.g., connection issues)
		// Log the error and proceed to fetch from repository, or return error based on policy
		fmt.Printf("Error getting data from cache (key: %s): %v. Fetching from repository.\n", cacheKey, err)
		// Depending on strictness, you might return err here.
	}

	// 2. Cache miss or other cache error, fetch from repository
	summaryCards, err := s.dashboardRepo.GetSummaryCardsData(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting summary cards: %w", err)
	}

	accountSummaries, err := s.dashboardRepo.GetAccountSummaries(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting account summaries: %w", err)
	}

	upcomingBills, err := s.dashboardRepo.GetUpcomingBills(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting upcoming bills: %w", err)
	}

	revenueExpenseChartData, err := s.dashboardRepo.GetRevenueExpenseChartData(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting revenue/expense chart data: %w", err)
	}

	expenseCategoryChartData, err := s.dashboardRepo.GetExpenseCategoryChartData(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting expense by category chart data: %w", err)
	}

	personalizedRecommendations, err := s.dashboardRepo.GetPersonalizedRecommendations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting personalized recommendations: %w", err)
	}

	// Assemble all fetched data into the DashboardSummary struct
	fetchedSummary := entity_dashboard.DashboardSummary{
		SummaryCards:              *summaryCards,
		AccountSummary:            accountSummaries,
		UpcomingBills:             upcomingBills,
		RevenueExpenseChart:       revenueExpenseChartData,
		ExpenseByCategoryChart:    expenseCategoryChartData,
		PersonalizedRecommendations: personalizedRecommendations,
	}

	// 3. Set data into cache
	if err := s.cache.Set(ctx, cacheKey, fetchedSummary, s.cacheTTL); err != nil {
		// Log caching error but don't fail the request
		fmt.Printf("Error setting data to cache (key: %s): %v\n", cacheKey, err)
	}

	return &fetchedSummary, nil
}
