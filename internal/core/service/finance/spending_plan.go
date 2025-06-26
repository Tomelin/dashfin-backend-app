package finance

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

const spendingPlanCacheTTL = 1 * time.Minute

// GetSpendingPlan retrieves a spending plan for a given user, using cache.
func (s *spendingPlanService) GetSpendingPlan(ctx context.Context, userID string) (*entity_finance.SpendingPlan, error) {
	cacheKey := fmt.Sprintf("spending_plan:%s", userID)
	// Try to get from cache
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil { // Found in cache
		var plan entity_finance.SpendingPlan
		if jsonErr := json.Unmarshal([]byte(cachedData), &plan); jsonErr == nil {
			return &plan, nil
		} 
	} else if err != cache.ErrNotFound { // Actual cache error
		// Log cache error and fall through to repository
		log.Printf("Cache error fetching spending plan for UserID %s: %v", userID, err)
	}

	// Cache miss or cache error, fetch from repository
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
func (s *spendingPlanService) UpdateSpendingPlan(ctx context.Context, planData *entity_finance.SpendingPlan) (*entity_finance.SpendingPlan, error) {

	if planData == nil {
		return nil, fmt.Errorf("planData cannot be nil")
	}

	if planData.UserID == "" {
		return nil, fmt.Errorf("UserID cannot be empty")
	}

	var existingPlan *entity_finance.SpendingPlan
	var err error

	cacheKey := fmt.Sprintf("spending_plan:%s", planData.UserID)

	dataByCache, _ := s.cache.Get(ctx, cacheKey)
	if dataByCache == "" {
		existingPlan, err = s.repo.GetSpendingPlanByUserID(ctx, planData.UserID)
		if err != nil {
			if err.Error() == "spendingPlan not found" {
				existingPlan, err = s.CreateSpendingPlan(ctx, planData)
				if err != nil {
					return nil, err
				}
				s.setCacheSpendingPlan(ctx, cacheKey, existingPlan)
				return existingPlan, nil
			}
			return nil, err
		}

		s.setCacheSpendingPlan(ctx, cacheKey, existingPlan)

		return existingPlan, nil
	}

	s.cache.Set(ctx, cacheKey, *planData, spendingPlanCacheTTL)

	err = json.Unmarshal([]byte(dataByCache), &existingPlan)
	if err != nil {
		return nil, err
	}

	err = s.repo.UpdateSpendingPlan(ctx, planData)
	if err != nil {
		return nil, err
	}

	return existingPlan, nil
}

// CreateSpendingPlan
func (s *spendingPlanService) CreateSpendingPlan(ctx context.Context, planData *entity_finance.SpendingPlan) (*entity_finance.SpendingPlan, error) {

	if planData == nil {
		return nil, fmt.Errorf("planData cannot be nil")
	}

	if planData.UserID == "" {
		return nil, fmt.Errorf("UserID cannot be empty")
	}

	response, err := s.repo.CreateSpendingPlan(ctx, planData)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *spendingPlanService) setCacheSpendingPlan(ctx context.Context, cacheKey string, planData *entity_finance.SpendingPlan) {

	s.cache.Set(ctx, cacheKey, *planData, spendingPlanCacheTTL)
}
