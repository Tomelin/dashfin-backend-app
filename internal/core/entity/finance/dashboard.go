package finance

import "time"

// Dashboard represents the data displayed on the main financial dashboard.
type Dashboard struct {
	SummaryCards                  SummaryCards                  `json:"summary_cards"`
	AccountSummaryData            []AccountSummary              `json:"account_summary_data"`
	UpcomingBillsData             []UpcomingBill                `json:"upcoming_bills_data"`
	RevenueExpenseChartData       []RevenueExpenseChartItem     `json:"revenue_expense_chart_data"`
	ExpenseCategoryChartData      []ExpenseCategoryChartItem    `json:"expense_category_chart_data"`
	PersonalizedRecommendationsData []PersonalizedRecommendation `json:"personalized_recommendations_data"`
}

// SummaryCards holds the data for the summary cards at the top of the dashboard.
type SummaryCards struct {
	TotalBalance    float64 `json:"totalBalance"`    // Saldo total consolidado de todas as contas (R$).
	MonthlyRevenue  float64 `json:"monthlyRevenue"`  // Total de receitas no mês corrente (R$).
	MonthlyExpenses float64 `json:"monthlyExpenses"` // Total de despesas no mês corrente (R$).
	GoalsProgress   string  `json:"goalsProgress"`   // Progresso geral das metas financeiras.
}

// AccountSummary represents a summary of a single account.
type AccountSummary struct {
	AccountName string  `json:"accountName"` // Nome da conta.
	Balance     float64 `json:"balance"`     // Saldo da conta.
}

// UpcomingBill represents a bill that is due soon.
type UpcomingBill struct {
	BillName string    `json:"billName"` // Nome da conta a pagar.
	Amount   float64   `json:"amount"`   // Valor da conta.
	DueDate  time.Time `json:"dueDate"`  // Data de vencimento (YYYY-MM-DD).
}

// RevenueExpenseChartItem represents a data point for the revenue vs. expenses chart.
type RevenueExpenseChartItem struct {
	Month    string  `json:"month"`    // Mês (e.g., "Jan", "Jan/24").
	Revenue  float64 `json:"revenue"`  // Total de receitas no mês.
	Expenses float64 `json:"expenses"` // Total de despesas no mês.
}

// ExpenseCategoryChartItem represents a data point for the expenses by category chart.
type ExpenseCategoryChartItem struct {
	Name  string  `json:"name"`  // Nome da categoria.
	Value float64 `json:"value"` // Valor gasto na categoria.
	// Fill string `json:"fill,omitempty"` // Cor para o gráfico (opcional).
}

// PersonalizedRecommendation represents a personalized financial recommendation.
type PersonalizedRecommendation struct {
	RecommendationID string `json:"recommendationId"` // ID da recomendação.
	Title            string `json:"title"`            // Título da recomendação.
	DescriptionText  string `json:"descriptionText"`  // Texto descritivo da recomendação.
	Category         string `json:"category"`         // Categoria da recomendação.
}
