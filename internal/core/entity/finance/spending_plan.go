package entity_finance

import "time"

// SpendingPlan represents a user's monthly spending plan.
type SpendingPlan struct {
	MonthlyIncome   float64          `json:"monthlyIncome"`
	CategoryBudgets []CategoryBudget `json:"categoryBudgets"`
	UserID          string           `json:"userId"` // Managed by the backend
	CreatedAt       time.Time        `json:"createdAt"` // Managed by the backend
	UpdatedAt       time.Time        `json:"updatedAt"` // Managed by the backend
}

// CategoryBudget represents the budget for a specific category within a spending plan.
type CategoryBudget struct {
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
}
