package dashboard

import (
	"context"
	"time"
)

// Dashboard represents the data displayed on the main financial dashboard.
type Dashboard struct {
	SummaryCards                    SummaryCards                 `json:"summary_cards"`
	AccountSummaryData              []AccountSummary             `json:"account_summary_data"`
	UpcomingBillsData               []UpcomingBill               `json:"upcoming_bills_data"`
	RevenueExpenseChartData         []RevenueExpenseChartItem    `json:"revenue_expense_chart_data"`
	ExpenseCategoryChartData        []ExpenseCategoryChartItem   `json:"expense_category_chart_data"`
	PersonalizedRecommendationsData []PersonalizedRecommendation `json:"personalized_recommendations_data"`
}

type UpcomingBillData struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Amount  float64 `json:"amount"`
	DueDate string  `json:"dueDate"`
}

type AccountBalanceItem struct {
	ID          string  `json:"id,omitempty"`
	AccountName string  `json:"accountName"`
	BankName    string  `json:"bankName"`
	Balance     float64 `json:"balance"`
	UserID      string  `json:"userId"`
}

type MonthlyFinancialSummaryItem struct {
	ID            string    `json:"id,omitempty"`
	Month         string    `json:"month"`         // Changed order to put ID first
	TotalIncome   float64   `json:"totalIncome"`   // Removed extra space
	TotalExpenses float64   `json:"totalExpenses"` // Removed extra space
	UserID        string    `json:"userId"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// SummaryCards holds the data for the summary cards at the top of the dashboard.
type SummaryCards struct {
	TotalBalance                 float64                       `json:"totalBalance"`    // Saldo total consolidado de todas as contas (R$).
	MonthlyRevenue               float64                       `json:"monthlyRevenue"`  // Total de receitas no mês corrente (R$).
	MonthlyExpenses              float64                       `json:"monthlyExpenses"` // Total de despesas no mês corrente (R$).
	GoalsProgress                string                        `json:"goalsProgress"`   // Progresso geral das metas financeiras.
	TotalBalanceChangePercent    float64                       `json:"totalBalanceChangePercent,omitempty"`
	MonthlyRevenueChangePercent  float64                       `json:"monthlyRevenueChangePercent,omitempty"`
	MonthlyExpensesChangePercent float64                       `json:"monthlyExpensesChangePercent,omitempty"`
	GoalsProgressDescription     string                        `json:"goalsProgressDescription,omitempty"`
	UpcomingBillsData            []UpcomingBillData            `json:"upcomingBillsData"`
	AccountBalances              []AccountBalanceItem          `json:"accountBalances,omitempty"`
	MonthlyFinancialSummary      []MonthlyFinancialSummaryItem `json:"monthlyFinancialSummary,omitempty"`
}

// AccountSummary represents a summary of a single account.
type AccountSummary struct {
	AccountName string  `json:"accountName"` // Nome da conta.
	Balance     float64 `json:"balance"`     // Saldo da conta.
}

// UpcomingBill represents a bill that is due soon.
type UpcomingBill struct {
	BillName string  `json:"billName"` // Nome da conta a pagar.
	Amount   float64 `json:"amount"`   // Valor da conta.
	DueDate  string  `json:"dueDate"`  // Data de vencimento (YYYY-MM-DD).
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

// DashboardRepositoryInterface defines the operations for storing and retrieving dashboard data.
// This would typically be for caching purposes.
type DashboardRepositoryInterface interface {
	// GetDashboard retrieves the dashboard data for a given user ID.
	// It returns the Dashboard, a boolean indicating if it was found (true if found, false if not),
	// and an error if any occurred (other than not found).
	GetDashboard(ctx context.Context, userID string) (dashboard *Dashboard, found bool, err error)

	// SaveDashboard stores the dashboard data for a given user ID.
	// A TTL (time-to-live) duration should be considered by the implementation for caching.
	SaveDashboard(ctx context.Context, userID string, dashboard *Dashboard, ttl time.Duration) error

	// DeleteDashboard explicitly removes dashboard data for a user, e.g., on logout or data reset.
	DeleteDashboard(ctx context.Context, userID string) error

	GetBankAccountBalanceByID(ctx context.Context, userID, bankName *string) (*AccountBalanceItem, error)
	UpdateBankAccountBalance(ctx context.Context, userID *string, data *AccountBalanceItem) error
	GetBankAccountBalance(ctx context.Context, userID *string) ([]AccountBalanceItem, error)
	UpdateFinancialSummary(ctx context.Context, userID *string, data *MonthlyFinancialSummaryItem) error
	GetFinancialSummary(ctx context.Context, userID *string) ([]MonthlyFinancialSummaryItem, error)
}
