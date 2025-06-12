// Package dashboard defines entities related to the dashboard feature.
package dashboard

// SummaryCards represents the summary cards section of the dashboard.
type SummaryCards struct {
	TotalBalance    float64 `json:"totalBalance"`    // Total balance across all accounts.
	TotalIncome     float64 `json:"totalIncome"`     // Total income received this month.
	TotalExpenses   float64 `json:"totalExpenses"`   // Total expenses incurred this month.
	SavingsGoal     float64 `json:"savingsGoal"`     // Monthly savings goal.
	UpcomingBills   int     `json:"upcomingBills"`   // Number of upcoming bills in the next 30 days.
	UncategorizedTx int     `json:"uncategorizedTx"` // Number of uncategorized transactions.
}

// AccountSummaryData represents the data for a single account in the account summary.
type AccountSummaryData struct {
	AccountID   string  `json:"accountID"`   // Unique identifier for the account.
	AccountName string  `json:"accountName"` // Name of the account (e.g., "Checking Account", "Savings Account").
	Balance     float64 `json:"balance"`     // Current balance of the account.
	BankName    string  `json:"bankName"`    // Name of the bank.
	AccountType string  `json:"accountType"` // Type of account (e.g., "checking", "savings", "credit card").
}

// UpcomingBill represents an upcoming bill.
type UpcomingBill struct {
	BillID      string  `json:"billID"`      // Unique identifier for the bill.
	Description string  `json:"description"` // Description of the bill (e.g., "Netflix Subscription", "Rent Payment").
	Amount      float64 `json:"amount"`      // Amount due for the bill.
	DueDate     string  `json:"dueDate"`     // Due date of the bill (YYYY-MM-DD).
	IsPaid      bool    `json:"isPaid"`      // Status of the bill (paid/unpaid).
	PayNowLink  string  `json:"payNowLink"`  // Link to pay the bill.
}

// RevenueExpenseChartData represents the data for the revenue vs. expense chart.
type RevenueExpenseChartData struct {
	Month   string  `json:"month"`   // Month for the data point (e.g., "Jan", "Feb").
	Revenue float64 `json:"revenue"` // Total revenue for the month.
	Expense float64 `json:"expense"` // Total expenses for the month.
}

// ExpenseCategoryChartData represents the data for the expense by category chart.
type ExpenseCategoryChartData struct {
	Category string  `json:"category"` // Expense category (e.g., "Groceries", "Utilities", "Entertainment").
	Amount   float64 `json:"amount"`   // Total amount spent in this category.
}

// PersonalizedRecommendation represents a personalized recommendation.
type PersonalizedRecommendation struct {
	RecommendationID string `json:"recommendationID"` // Unique identifier for the recommendation.
	Title            string `json:"title"`            // Title of the recommendation.
	Description      string `json:"description"`      // Detailed description of the recommendation.
	ActionLink       string `json:"actionLink"`       // Link to act on the recommendation.
	Priority         string `json:"priority"`         // Priority of the recommendation (e.g., "high", "medium", "low").
}

// DashboardSummary is the top-level struct that holds all sections of the dashboard summary.
type DashboardSummary struct {
	SummaryCards              SummaryCards                 `json:"summaryCards"`              // Summary cards section.
	AccountSummary            []AccountSummaryData         `json:"accountSummary"`            // List of account summaries.
	UpcomingBills             []UpcomingBill               `json:"upcomingBills"`             // List of upcoming bills.
	RevenueExpenseChart       []RevenueExpenseChartData    `json:"revenueExpenseChart"`       // Data for the revenue vs. expense chart.
	ExpenseByCategoryChart    []ExpenseCategoryChartData   `json:"expenseByCategoryChart"`    // Data for the expense by category chart.
	PersonalizedRecommendations []PersonalizedRecommendation `json:"personalizedRecommendations"` // List of personalized recommendations.
}
