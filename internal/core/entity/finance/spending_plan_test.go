package entity_finance_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
)

func TestSpendingPlan_Instantiation(t *testing.T) {
	expectedMonthlyIncome := 5000.0
	expectedCategory := "Groceries"
	expectedAmount := 500.0
	expectedPercentage := 0.1

	categoryBudgets := []entity_finance.CategoryBudget{
		{
			Category:   expectedCategory,
			Amount:     expectedAmount,
			Percentage: expectedPercentage,
		},
		{
			Category:   "Transport",
			Amount:     300.0,
			Percentage: 0.06,
		},
	}

	spendingPlan := entity_finance.SpendingPlan{
		MonthlyIncome:   expectedMonthlyIncome,
		CategoryBudgets: categoryBudgets,
		UserID:          "user123",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	assert.NotNil(t, spendingPlan)
	assert.Equal(t, expectedMonthlyIncome, spendingPlan.MonthlyIncome)
	assert.Len(t, spendingPlan.CategoryBudgets, 2)
	assert.Equal(t, expectedCategory, spendingPlan.CategoryBudgets[0].Category)
	assert.Equal(t, expectedAmount, spendingPlan.CategoryBudgets[0].Amount)
	assert.Equal(t, expectedPercentage, spendingPlan.CategoryBudgets[0].Percentage)
	assert.Equal(t, "user123", spendingPlan.UserID)
	assert.NotZero(t, spendingPlan.CreatedAt)
	assert.NotZero(t, spendingPlan.UpdatedAt)
}

func TestCategoryBudget_Instantiation(t *testing.T) {
	expectedCategory := "Entertainment"
	expectedAmount := 200.0
	expectedPercentage := 0.04

	categoryBudget := entity_finance.CategoryBudget{
		Category:   expectedCategory,
		Amount:     expectedAmount,
		Percentage: expectedPercentage,
	}

	assert.NotNil(t, categoryBudget)
	assert.Equal(t, expectedCategory, categoryBudget.Category)
	assert.Equal(t, expectedAmount, categoryBudget.Amount)
	assert.Equal(t, expectedPercentage, categoryBudget.Percentage)
}
