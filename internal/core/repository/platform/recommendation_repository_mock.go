// Package platform provides mock implementations of repository interfaces.
package platform

import (
	"context"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
)

// MockRecommendationRepository is a mock implementation of RecommendationRepositoryInterface.
type MockRecommendationRepository struct{}

// GetPersonalizedRecommendationsByUserID returns a fixed list of recommendations for testing.
func (m *MockRecommendationRepository) GetPersonalizedRecommendationsByUserID(ctx context.Context, userID string, limit int) ([]entity_dashboard.PersonalizedRecommendation, error) {
	recs := []entity_dashboard.PersonalizedRecommendation{
		{
			RecommendationID: "rec_1",
			Title:            "Setup Automatic Savings",
			Description:      "Automate your savings to reach your goals faster. You can set up a recurring transfer to your savings account.",
			ActionLink:       "http://example.com/setup-auto-save",
			Priority:         "high",
		},
		{
			RecommendationID: "rec_2",
			Title:            "Review Your Monthly Subscriptions",
			Description:      "You are subscribed to 5 services. Review them to see if you still need all of them.",
			ActionLink:       "http://example.com/review-subscriptions",
			Priority:         "medium",
		},
		{
			RecommendationID: "rec_3",
			Title:            "Consider a High-Yield Savings Account",
			Description:      "Earn more interest on your savings by switching to a high-yield account.",
			ActionLink:       "http://example.com/explore-hysa",
			Priority:         "medium",
		},
	}
	if limit > 0 && len(recs) > limit {
		return recs[:limit], nil
	}
	return recs, nil
}

// NewMockRecommendationRepository creates a new instance of MockRecommendationRepository.
func NewMockRecommendationRepository() *MockRecommendationRepository {
	return &MockRecommendationRepository{}
}
