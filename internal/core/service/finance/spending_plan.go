package service_finance

import (
	"context"
	"encoding/json"
	"fmt"
	"log" // For logging cache errors
	"time"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	"github.com/Tomelin/dashfin-backend-app/pkg/cache" // Import cache package
)

// spendingPlanService implements the SpendingPlanService interface.
type spendingPlanService struct {
	repo  entity_finance.SpendingPlanRepositoryInterface
	cache cache.CacheService // Added cache field
}

// InitializeSpendingPlanService creates a new SpendingPlanService.
func InitializeSpendingPlanService(repo entity_finance.SpendingPlanRepositoryInterface, cacheService cache.CacheService) (entity_finance.SpendingPlanServiceInterface, error) {

	if repo == nil {
		return nil, fmt.Errorf("spendingPlanRepository cannot be nil")
	}

	if cacheService == nil {
		return nil, fmt.Errorf("cacheService cannot be nil")
	}

	return &spendingPlanService{repo: repo, cache: cacheService}, nil
}

const spendingPlanCacheTTL = 10 * time.Minute

// GetSpendingPlan retrieves a spending plan for a given user, using cache.
func (s *spendingPlanService) GetSpendingPlan(ctx context.Context, userID string) (*entity_finance.SpendingPlan, error) {
	cacheKey := fmt.Sprintf("spending_plan:%s", userID)
	log.Println(" cacheKey", cacheKey)
	// Try to get from cache
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil { // Found in cache
		var plan entity_finance.SpendingPlan
		if jsonErr := json.Unmarshal([]byte(cachedData), &plan); jsonErr == nil {
			return &plan, nil
		} else {
			// Log unmarshal error and fall through to repository
			log.Printf("Error unmarshalling cached spending plan for UserID %s: %v", userID, jsonErr)
		}
	} else if err != cache.ErrNotFound { // Actual cache error
		// Log cache error and fall through to repository
		log.Printf("Cache error fetching spending plan for UserID %s: %v", userID, err)
	}

	// Cache miss or cache error, fetch from repository
	log.Println("Service userID", userID)
	plan, repoErr := s.repo.GetSpendingPlanByUserID(ctx, userID)
	if repoErr != nil {
		return nil, repoErr
	}

	if plan != nil { // Found in repo, store in cache
		jsonData, jsonErr := json.Marshal(plan)
		if jsonErr == nil {
			if cacheSetErr := s.cache.Set(ctx, cacheKey, string(jsonData), spendingPlanCacheTTL); cacheSetErr != nil {
				log.Printf("Error setting cache for spending plan UserID %s: %v", userID, cacheSetErr)
			}
		} else {
			log.Printf("Error marshalling spending plan for caching UserID %s: %v", userID, jsonErr)
		}
	}
	return plan, nil
}

// SaveSpendingPlan creates or updates a spending plan for a given user and invalidates cache.
func (s *spendingPlanService) SaveSpendingPlan(ctx context.Context, planData *entity_finance.SpendingPlan) (*entity_finance.SpendingPlan, error) {

	if planData == nil {
		return nil, fmt.Errorf("planData cannot be nil")
	}

	if planData.UserID == "" {
		return nil, fmt.Errorf("UserID cannot be empty")
	}

	cacheKey := fmt.Sprintf("spending_plan:%s", planData.UserID)

	existingPlan, err := s.repo.GetSpendingPlanByUserID(ctx, planData.UserID)
	if err != nil {
		// Assuming error means "not found" for simplicity here. (As per previous logic)
		// This means we are creating a new plan.
		// A more robust solution would check the specific error type.
		newPlan := &entity_finance.SpendingPlan{
			UserID:          planData.UserID,
			MonthlyIncome:   planData.MonthlyIncome,
			CategoryBudgets: planData.CategoryBudgets,
			CreatedAt:       time.Now(), // Service sets this
			UpdatedAt:       time.Now(), // Service sets this
		}
		// UserID from context takes precedence if planData.UserID is different or empty
		newPlan.UserID = planData.UserID

		if repoErr := s.repo.SaveSpendingPlan(ctx, newPlan); repoErr != nil {
			return nil, repoErr
		}
		// Invalidate cache
		if cacheDelErr := s.cache.Delete(ctx, cacheKey); cacheDelErr != nil && cacheDelErr != cache.ErrNotFound {
			log.Printf("Error deleting cache for spending plan UserID %s after create: %v", planData.UserID, cacheDelErr)
		}
		return newPlan, nil // Return the newPlan that was successfully saved
	}

	// Plan exists, update it
	existingPlan.MonthlyIncome = planData.MonthlyIncome
	existingPlan.CategoryBudgets = planData.CategoryBudgets
	existingPlan.UpdatedAt = time.Now() // Service updates this
	// Ensure UserID is consistent if it came from planData (though userID from context is authoritative)
	existingPlan.UserID = planData.UserID

	if repoErr := s.repo.SaveSpendingPlan(ctx, existingPlan); repoErr != nil {
		return nil, repoErr
	}

	// Invalidate cache
	if cacheDelErr := s.cache.Delete(ctx, cacheKey); cacheDelErr != nil && cacheDelErr != cache.ErrNotFound {
		log.Printf("Error deleting cache for spending plan UserID %s after update: %v", planData.UserID, cacheDelErr)
	}
	return existingPlan, nil // Return the existingPlan that was successfully updated
}
