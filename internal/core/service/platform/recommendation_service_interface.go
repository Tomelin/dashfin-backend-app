// Package platform defines service interfaces for platform-related business logic.
package platform

import (
	"context"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
)

// RecommendationServiceInterface defines methods for recommendation-related business logic.
type RecommendationServiceInterface interface {
	// GetPersonalizedRecommendationsByUserID retrieves a list of personalized recommendations for a given user.
	GetPersonalizedRecommendationsByUserID(ctx context.Context, userID string, limit int) ([]entity_dashboard.PersonalizedRecommendation, error)
}
