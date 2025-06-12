// Package dashboard_test contains tests for the dashboard service.
package dashboard_test

import (
	"context"
	"errors"
	"testing"
	"time"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
	repo_dashboard_mocks "github.com/Tomelin/dashfin-backend-app/internal/core/repository/dashboard/mocks"
	"github.com/Tomelin/dashfin-backend-app/internal/core/service/dashboard"
	"github.com/Tomelin/dashfin-backend-app/pkg/cache"
	cache_mocks "github.com/Tomelin/dashfin-backend-app/pkg/cache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDashboardService_GetDashboardSummary(t *testing.T) {
	ctx := context.Background()
	userID := "test-user-123"
	cacheTTLSeconds := 300
	cacheTTL := time.Duration(cacheTTLSeconds) * time.Second

	mockRepo := new(repo_dashboard_mocks.MockDashboardRepository)
	mockCache := new(cache_mocks.MockRedisCache)

	service := dashboard.NewDashboardService(mockRepo, mockCache, cacheTTLSeconds)

	// Expected data from repository
	expectedSummaryCards := &entity_dashboard.SummaryCards{TotalBalance: 1000}
	expectedAccountSummaries := []entity_dashboard.AccountSummaryData{{AccountName: "Test Account"}}
	expectedUpcomingBills := []entity_dashboard.UpcomingBill{{Description: "Test Bill"}}
	expectedRevenueExpenseChart := []entity_dashboard.RevenueExpenseChartData{{Month: "Jan", Revenue: 100, Expense: 50}}
	expectedExpenseCategoryChart := []entity_dashboard.ExpenseCategoryChartData{{Category: "Food", Amount: 20}}
	expectedRecommendations := []entity_dashboard.PersonalizedRecommendation{{Title: "Test Rec"}}

	expectedFullSummary := &entity_dashboard.DashboardSummary{
		SummaryCards:              *expectedSummaryCards,
		AccountSummary:            expectedAccountSummaries,
		UpcomingBills:             expectedUpcomingBills,
		RevenueExpenseChart:       expectedRevenueExpenseChart,
		ExpenseByCategoryChart:    expectedExpenseCategoryChart,
		PersonalizedRecommendations: expectedRecommendations,
	}
	cacheKey := "dashboard_summary:" + userID

	t.Run("Cache Miss & Successful Repository Fetch", func(t *testing.T) {
		// Reset mocks for this sub-test if using shared mock instances that accumulate calls.
		// However, testify mocks are typically reset per On/AssertExpectations.
		// For clarity, one might re-initialize mocks per subtest or use t.Cleanup.
		currentMockRepo := new(repo_dashboard_mocks.MockDashboardRepository)
		currentMockCache := new(cache_mocks.MockRedisCache)
		currentService := dashboard.NewDashboardService(currentMockRepo, currentMockCache, cacheTTLSeconds)


		// Setup: Cache Miss
		currentMockCache.OnGetCacheMiss(ctx, cacheKey, mock.AnythingOfType("*entity_dashboard.DashboardSummary"))

		// Setup: Repository Success
		currentMockRepo.On("GetSummaryCardsData", ctx, userID).Return(expectedSummaryCards, nil)
		currentMockRepo.On("GetAccountSummaries", ctx, userID).Return(expectedAccountSummaries, nil)
		currentMockRepo.On("GetUpcomingBills", ctx, userID).Return(expectedUpcomingBills, nil)
		currentMockRepo.On("GetRevenueExpenseChartData", ctx, userID).Return(expectedRevenueExpenseChart, nil)
		currentMockRepo.On("GetExpenseCategoryChartData", ctx, userID).Return(expectedExpenseCategoryChart, nil)
		currentMockRepo.On("GetPersonalizedRecommendations", ctx, userID).Return(expectedRecommendations, nil)

		// Setup: Cache Set Expectation
		currentMockCache.OnSet(ctx, cacheKey, *expectedFullSummary, cacheTTL, nil) // Expect the non-pointer summary

		// Action
		summary, err := currentService.GetDashboardSummary(ctx, userID)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, expectedFullSummary, summary)

		currentMockCache.AssertExpectations(t)
		currentMockRepo.AssertExpectations(t)
	})

	t.Run("Cache Hit", func(t *testing.T) {
		currentMockRepo := new(repo_dashboard_mocks.MockDashboardRepository) // Fresh mock
		currentMockCache := new(cache_mocks.MockRedisCache)   // Fresh mock
		currentService := dashboard.NewDashboardService(currentMockRepo, currentMockCache, cacheTTLSeconds)

		// Setup: Cache Hit
		// The OnGetSuccess helper populates the passed *entity_dashboard.DashboardSummary
		currentMockCache.OnGetSuccess(ctx, cacheKey, mock.AnythingOfType("*entity_dashboard.DashboardSummary"), *expectedFullSummary)

		// Action
		summary, err := currentService.GetDashboardSummary(ctx, userID)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, expectedFullSummary, summary)

		currentMockCache.AssertCalled(t, "Get", ctx, cacheKey, mock.AnythingOfType("*entity_dashboard.DashboardSummary"))
		currentMockCache.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		// Verify repository methods were NOT called
		currentMockRepo.AssertNotCalled(t, "GetSummaryCardsData", mock.Anything, mock.Anything)
		// ... (assert NotCalled for other repo methods too for completeness, though if one isn't called, others likely aren't)
	})

	t.Run("Error from Repository", func(t *testing.T) {
		currentMockRepo := new(repo_dashboard_mocks.MockDashboardRepository)
		currentMockCache := new(cache_mocks.MockRedisCache)
		currentService := dashboard.NewDashboardService(currentMockRepo, currentMockCache, cacheTTLSeconds)

		repoError := errors.New("repository error")

		// Setup: Cache Miss
		currentMockCache.OnGetCacheMiss(ctx, cacheKey, mock.AnythingOfType("*entity_dashboard.DashboardSummary"))
		// Setup: Repository Error (e.g., from GetSummaryCardsData)
		currentMockRepo.On("GetSummaryCardsData", ctx, userID).Return(nil, repoError)
		// No need to mock other repo calls if the first one fails

		// Action
		summary, err := currentService.GetDashboardSummary(ctx, userID)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, summary)
		assert.Contains(t, err.Error(), repoError.Error()) // Check if original error is wrapped

		currentMockCache.AssertCalled(t, "Get", ctx, cacheKey, mock.AnythingOfType("*entity_dashboard.DashboardSummary"))
		currentMockRepo.AssertCalled(t, "GetSummaryCardsData", ctx, userID)
		currentMockCache.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Error from Cache Get (not ErrCacheMiss)", func(t *testing.T) {
		currentMockRepo := new(repo_dashboard_mocks.MockDashboardRepository)
		currentMockCache := new(cache_mocks.MockRedisCache)
		currentService := dashboard.NewDashboardService(currentMockRepo, currentMockCache, cacheTTLSeconds)

		cacheGetError := errors.New("unexpected cache read error")

		// Setup: Cache Get returns an unexpected error
		currentMockCache.OnGetError(ctx, cacheKey, mock.AnythingOfType("*entity_dashboard.DashboardSummary"), cacheGetError)

		// If cache read fails unexpectedly, service might try to fetch from repo.
		// This depends on the service's error handling logic for cache.Get
		// The current service implementation logs the error and proceeds to repo.
		currentMockRepo.On("GetSummaryCardsData", ctx, userID).Return(expectedSummaryCards, nil)
		currentMockRepo.On("GetAccountSummaries", ctx, userID).Return(expectedAccountSummaries, nil)
		currentMockRepo.On("GetUpcomingBills", ctx, userID).Return(expectedUpcomingBills, nil)
		currentMockRepo.On("GetRevenueExpenseChartData", ctx, userID).Return(expectedRevenueExpenseChart, nil)
		currentMockRepo.On("GetExpenseCategoryChartData", ctx, userID).Return(expectedExpenseCategoryChart, nil)
		currentMockRepo.On("GetPersonalizedRecommendations", ctx, userID).Return(expectedRecommendations, nil)
		currentMockCache.OnSet(ctx, cacheKey, *expectedFullSummary, cacheTTL, nil)


		// Action
		summary, err := currentService.GetDashboardSummary(ctx, userID)

		// Assertions
		// The service currently logs cache errors and proceeds. So, no error returned to client.
		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, expectedFullSummary, summary)

		currentMockCache.AssertExpectations(t)
		currentMockRepo.AssertExpectations(t)
	})

	t.Run("Error from Cache Set", func(t *testing.T) {
		currentMockRepo := new(repo_dashboard_mocks.MockDashboardRepository)
		currentMockCache := new(cache_mocks.MockRedisCache)
		currentService := dashboard.NewDashboardService(currentMockRepo, currentMockCache, cacheTTLSeconds)

		cacheSetError := errors.New("cache write error")

		// Setup: Cache Miss
		currentMockCache.OnGetCacheMiss(ctx, cacheKey, mock.AnythingOfType("*entity_dashboard.DashboardSummary"))

		// Setup: Repository Success
		currentMockRepo.On("GetSummaryCardsData", ctx, userID).Return(expectedSummaryCards, nil)
		currentMockRepo.On("GetAccountSummaries", ctx, userID).Return(expectedAccountSummaries, nil)
		currentMockRepo.On("GetUpcomingBills", ctx, userID).Return(expectedUpcomingBills, nil)
		currentMockRepo.On("GetRevenueExpenseChartData", ctx, userID).Return(expectedRevenueExpenseChart, nil)
		currentMockRepo.On("GetExpenseCategoryChartData", ctx, userID).Return(expectedExpenseCategoryChart, nil)
		currentMockRepo.On("GetPersonalizedRecommendations", ctx, userID).Return(expectedRecommendations, nil)

		// Setup: Cache Set returns an error
		currentMockCache.OnSet(ctx, cacheKey, *expectedFullSummary, cacheTTL, cacheSetError)

		// Action
		summary, err := currentService.GetDashboardSummary(ctx, userID)

		// Assertions
		// Cache set failure should not fail the main operation
		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, expectedFullSummary, summary)

		currentMockCache.AssertExpectations(t)
		currentMockRepo.AssertExpectations(t)
	})
}
