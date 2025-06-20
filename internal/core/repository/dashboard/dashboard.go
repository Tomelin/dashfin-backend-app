package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	dashboardEntity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
	"github.com/Tomelin/dashfin-backend-app/internal/core/repository"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

// InMemoryDashboardRepository implements dashboardEntity.DashboardRepositoryInterface using an in-memory store.
// This is suitable for single-instance deployments or testing. For multi-instance, a distributed cache (e.g., Redis) would be needed.
type InMemoryDashboardRepository struct {
	store      map[string]cachedDashboardItem
	mu         sync.RWMutex // To make operations safe for concurrent use
	collection string
	db         database.FirebaseDBInterface
}

type cachedDashboardItem struct {
	dashboard *dashboardEntity.Dashboard
	expiresAt time.Time
}

// Ensure InMemoryDashboardRepository implements the interface (compile-time check)
var _ dashboardEntity.DashboardRepositoryInterface = (*InMemoryDashboardRepository)(nil)

// NewInMemoryDashboardRepository creates a new InMemoryDashboardRepository.
func NewInMemoryDashboardRepository(db database.FirebaseDBInterface) *InMemoryDashboardRepository {
	return &InMemoryDashboardRepository{
		store:      make(map[string]cachedDashboardItem),
		collection: "dashboard",
		db:         db,
	}
}

// GetDashboard retrieves the dashboard data for a given user ID from the in-memory store.
func (r *InMemoryDashboardRepository) GetDashboard(ctx context.Context, userID string) (*dashboardEntity.Dashboard, bool, error) {
	if userID == "" {
		return nil, false, fmt.Errorf("userID cannot be empty")
	}

	r.mu.RLock()
	item, found := r.store[userID]
	r.mu.RUnlock()

	if !found {
		return nil, false, nil // Not found, no error
	}

	// Check if the item has expired
	if time.Now().After(item.expiresAt) {
		// Item has expired, treat as not found and remove it
		r.mu.Lock()
		delete(r.store, userID)
		r.mu.Unlock()
		return nil, false, nil // Expired, treat as not found
	}

	// Item found and not expired
	// Return a copy to avoid external modification of the cached item if Dashboard is mutable
	// For now, returning direct pointer. Consider deep copy if Dashboard struct fields are pointers/slices that could be modified.
	return item.dashboard, true, nil
}

// SaveDashboard stores the dashboard data for a given user ID in the in-memory store with a TTL.
func (r *InMemoryDashboardRepository) SaveDashboard(ctx context.Context, userID string, dashboard *dashboardEntity.Dashboard, ttl time.Duration) error {
	if userID == "" {
		return fmt.Errorf("userID cannot be empty")
	}
	if dashboard == nil {
		return fmt.Errorf("dashboard cannot be nil")
	}
	if ttl <= 0 {
		return fmt.Errorf("ttl must be a positive duration")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.store[userID] = cachedDashboardItem{
		dashboard: dashboard, // Storing pointer directly
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

// DeleteDashboard removes dashboard data for a user from the in-memory store.
func (r *InMemoryDashboardRepository) DeleteDashboard(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("userID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.store, userID)
	return nil
}

func (r *InMemoryDashboardRepository) UpdateBankAccountBalance(ctx context.Context, userID *string, data *dashboardEntity.AccountBalanceItem) error {
	if data == nil {
		return errors.New("data is nil")
	}

	toMap, _ := utils.StructToMap(data)

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return err
	}

	if collection == nil || *collection == "" {
		return fmt.Errorf("%s collection is empty", r.collection)
	}

	toMap["type"] = "accountBalance"

	err = r.db.Update(ctx, *userID, toMap, *collection)
	if err != nil {
		return err
	}

	return nil
}

func (r *InMemoryDashboardRepository) GetBankAccountBalanceByID(ctx context.Context, userID, bankName *string) (*dashboardEntity.AccountBalanceItem, error) {
	if userID == nil || *userID == "" {
		return nil, errors.New("id is empty")
	}

	if bankName == nil || *bankName == "" {
		return nil, errors.New("bankName is empty")
	}

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, err
	}

	if collection == nil || *collection == "" {
		return nil, fmt.Errorf("%s collection is empty", r.collection)
	}

	filters := map[string]interface{}{
		"userID":   *userID,
		"bankName": *bankName,
		"type":     "accountBalance",
	}

	result, err := r.db.GetByFilter(ctx, filters, *collection)
	log.Println(result, err)
	if err != nil {
		return nil, err
	}

	var items []dashboardEntity.AccountBalanceItem
	if err := json.Unmarshal(result, &items); err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, errors.New("bank account not found")
	}

	return &items[0], nil
}

func (r *InMemoryDashboardRepository) GetBankAccountBalance(ctx context.Context, userID *string) ([]dashboardEntity.AccountBalanceItem, error) {
	if userID == nil || *userID == "" {
		return nil, errors.New("id is empty")
	}

	collection, err := repository.SetCollection(ctx, r.collection)
	log.Println("collection > ", collection)
	if err != nil {
		return nil, err
	}

	if collection == nil || *collection == "" {
		return nil, fmt.Errorf("%s collection is empty", r.collection)
	}

	filters := map[string]interface{}{
		"userId": *userID,
		"type":   "accountBalance",
	}

	log.Println("filters > ", filters)
	result, err := r.db.GetByFilter(ctx, filters, *collection)
	log.Println("result > ", result, err)
	if err != nil {
		return nil, err
	}

	var items []dashboardEntity.AccountBalanceItem
	if err := json.Unmarshal(result, &items); err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, errors.New("bank account not found")
	}

	return items, nil
}
