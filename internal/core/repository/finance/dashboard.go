package finance

import (
	"context"
	"fmt"
	"sync"
	"time"

	financeEntity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
)

// InMemoryDashboardRepository implements financeEntity.DashboardRepositoryInterface using an in-memory store.
// This is suitable for single-instance deployments or testing. For multi-instance, a distributed cache (e.g., Redis) would be needed.
type InMemoryDashboardRepository struct {
	store map[string]cachedDashboardItem
	mu    sync.RWMutex // To make operations safe for concurrent use
}

type cachedDashboardItem struct {
	dashboard  *financeEntity.Dashboard
	expiresAt time.Time
}

// NewInMemoryDashboardRepository creates a new InMemoryDashboardRepository.
func NewInMemoryDashboardRepository() *InMemoryDashboardRepository {
	return &InMemoryDashboardRepository{
		store: make(map[string]cachedDashboardItem),
	}
}

// GetDashboard retrieves the dashboard data for a given user ID from the in-memory store.
func (r *InMemoryDashboardRepository) GetDashboard(ctx context.Context, userID string) (*financeEntity.Dashboard, bool, error) {
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
func (r *InMemoryDashboardRepository) SaveDashboard(ctx context.Context, userID string, dashboard *financeEntity.Dashboard, ttl time.Duration) error {
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

// Ensure InMemoryDashboardRepository implements the interface (compile-time check)
var _ financeEntity.DashboardRepositoryInterface = (*InMemoryDashboardRepository)(nil)
