// Package platform provides mock implementations of service interfaces.
package platform

import (
	"context"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
)

// MockRecommendationService is a mock implementation of RecommendationServiceInterface.
type MockRecommendationService struct{}

// GetPersonalizedRecommendationsByUserID returns a fixed list of recommendations for testing.
// This mock service might provide slightly different or augmented data.
func (m *MockRecommendationService) GetPersonalizedRecommendationsByUserID(ctx context.Context, userID string, limit int) ([]entity_dashboard.PersonalizedRecommendation, error) {
	recs := []entity_dashboard.PersonalizedRecommendation{
		{
			RecommendationID: "serv_rec_1",
			Title:            "Service: Consolidate Your Debt",
			Description:      "Consolidating your high-interest debts could save you money on interest payments. Explore your options.",
			ActionLink:       "http://example.com/service/consolidate-debt",
			Priority:         "high",
		},
		{
			RecommendationID: "serv_rec_2",
			Title:            "Service: Create a Budget for Groceries",
			Description:      "Tracking your grocery spending can help you identify savings opportunities. Try our budgeting tool.",
			ActionLink:       "http://example.com/service/grocery-budget",
			Priority:         "medium",
		},
	}
	if limit > 0 && len(recs) > limit {
		return recs[:limit], nil
	}
	return recs, nil
}

// NewMockRecommendationService creates a new instance of MockRecommendationService.
func NewMockRecommendationService() *MockRecommendationService {
	return &MockRecommendationService{}
}
