// Package platform defines repository interfaces for platform-related data.
package platform

import (
	"context"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
)

// RecommendationRepositoryInterface defines methods for interacting with recommendation data.
type RecommendationRepositoryInterface interface {
	// GetPersonalizedRecommendationsByUserID retrieves a list of personalized recommendations for a given user.
	GetPersonalizedRecommendationsByUserID(ctx context.Context, userID string, limit int) ([]entity_dashboard.PersonalizedRecommendation, error)
}
