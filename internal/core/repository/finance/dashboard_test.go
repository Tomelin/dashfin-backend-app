package finance

import (
	"context"
	"testing"
	"time"

	financeEntity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryDashboardRepository_SaveAndGetDashboard_Nominal(t *testing.T) {
	repo := NewInMemoryDashboardRepository()
	ctx := context.Background()
	userID := "user-123"
	dashboardToSave := &financeEntity.Dashboard{
		SummaryCards: financeEntity.SummaryCards{TotalBalance: 1000},
	}
	ttl := 100 * time.Millisecond

	// Save
	err := repo.SaveDashboard(ctx, userID, dashboardToSave, ttl)
	assert.NoError(t, err)

	// Get immediately
	retrievedDashboard, found, err := repo.GetDashboard(ctx, userID)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.NotNil(t, retrievedDashboard)
	assert.Equal(t, dashboardToSave.SummaryCards.TotalBalance, retrievedDashboard.SummaryCards.TotalBalance)

	// Wait for TTL to expire
	time.Sleep(ttl + 50*time.Millisecond)

	// Get after expiration
	expiredDashboard, found, err := repo.GetDashboard(ctx, userID)
	assert.NoError(t, err)
	assert.False(t, found, "Dashboard should be expired and thus not found")
	assert.Nil(t, expiredDashboard, "Expired dashboard should be nil")
}

func TestInMemoryDashboardRepository_GetDashboard_NotFound(t *testing.T) {
	repo := NewInMemoryDashboardRepository()
	ctx := context.Background()
	userID := "user-nonexistent"

	retrievedDashboard, found, err := repo.GetDashboard(ctx, userID)
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, retrievedDashboard)
}

func TestInMemoryDashboardRepository_SaveDashboard_Validations(t *testing.T) {
	repo := NewInMemoryDashboardRepository()
	ctx := context.Background()
	userID := "user-validations"
	dashboard := &financeEntity.Dashboard{}

	err := repo.SaveDashboard(ctx, "", dashboard, 1*time.Second)
	assert.Error(t, err, "Should error on empty userID")

	err = repo.SaveDashboard(ctx, userID, nil, 1*time.Second)
	assert.Error(t, err, "Should error on nil dashboard")

	err = repo.SaveDashboard(ctx, userID, dashboard, 0*time.Second)
	assert.Error(t, err, "Should error on non-positive TTL (0)")

	err = repo.SaveDashboard(ctx, userID, dashboard, -1*time.Second)
	assert.Error(t, err, "Should error on non-positive TTL (-1)")
}

func TestInMemoryDashboardRepository_GetDashboard_EmptyUserID(t *testing.T) {
	repo := NewInMemoryDashboardRepository()
	ctx := context.Background()

	_, found, err := repo.GetDashboard(ctx, "")
	assert.Error(t, err, "Should error on empty userID for GetDashboard")
	assert.False(t, found)
}


func TestInMemoryDashboardRepository_DeleteDashboard(t *testing.T) {
	repo := NewInMemoryDashboardRepository()
	ctx := context.Background()
	userID := "user-to-delete"
	dashboardToSave := &financeEntity.Dashboard{
		SummaryCards: financeEntity.SummaryCards{TotalBalance: 2000},
	}
	ttl := 1 * time.Minute

	// Save
	err := repo.SaveDashboard(ctx, userID, dashboardToSave, ttl)
	assert.NoError(t, err)

	// Get to confirm it's there
	_, found, err := repo.GetDashboard(ctx, userID)
	assert.NoError(t, err)
	assert.True(t, found, "Dashboard should be found before delete")

	// Delete
	err = repo.DeleteDashboard(ctx, userID)
	assert.NoError(t, err)

	// Get again to confirm it's gone
	_, found, err = repo.GetDashboard(ctx, userID)
	assert.NoError(t, err)
	assert.False(t, found, "Dashboard should not be found after delete")

	// Test deleting non-existent item
	err = repo.DeleteDashboard(ctx, "user-nonexistent-delete")
	assert.NoError(t, err, "Deleting non-existent item should not error")

	// Test deleting with empty userID
	err = repo.DeleteDashboard(ctx, "")
	assert.Error(t, err, "Deleting with empty userID should error")
}

func TestInMemoryDashboardRepository_Concurrency(t *testing.T) {
	repo := NewInMemoryDashboardRepository()
	ctx := context.Background()
	userID := "concurrent-user"
	dashboardToSave := &financeEntity.Dashboard{ // This specific instance isn't used in the loops directly
		SummaryCards: financeEntity.SummaryCards{TotalBalance: 3000},
	}
	ttl := 200 * time.Millisecond // Longer TTL for concurrent test

	// Concurrent saves and gets
	done := make(chan bool)
	writerCount := 100
	readerCount := 100

	// Writer goroutine
	go func() {
		for i := 0; i < writerCount; i++ {
			val := float64(3000 + i)
			currentDashboard := &financeEntity.Dashboard{
				SummaryCards: financeEntity.SummaryCards{TotalBalance: val},
			}
			repo.SaveDashboard(ctx, userID, currentDashboard, ttl) // Use currentDashboard
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Reader goroutine
	var lastReadValue float64
	var readsWithValue int // Count how many times we actually read a value
	go func() {
		for i := 0; i < readerCount; i++ {
			dash, found, _ := repo.GetDashboard(ctx, userID)
			if found && dash != nil {
				lastReadValue = dash.SummaryCards.TotalBalance
				readsWithValue++
			}
			time.Sleep(2 * time.Millisecond)
		}
		done <- true
	}()

	<-done
	<-done

	assert.True(t, readsWithValue > 0, "Should have read a value at least once during concurrent access")
	// Asserting a specific lastReadValue can be flaky due to timing.
	// Instead, we check if *any* valid value was read.
	// If readsWithValue > 0, then lastReadValue was set from a valid dashboard.
	if readsWithValue > 0 {
		assert.True(t, lastReadValue >= 3000 && lastReadValue < float64(3000+writerCount), "A valid dashboard value should have been read concurrently")
	}


	// Check expiration after concurrent operations
	time.Sleep(ttl + 50*time.Millisecond)
	_, found, err := repo.GetDashboard(ctx, userID)
	assert.NoError(t, err)
	assert.False(t, found, "Dashboard should expire after concurrent tests")
}
