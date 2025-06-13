package dashboard

import "time"

// PlannedVsActualCategory represents the comparison of planned versus actual expenditure
// for a specific category within a given period.
// It includes the planned amount, actual amount spent, and the percentage of the planned amount that has been spent.
type PlannedVsActualCategory struct {
	// Category is the unique key for the expense category (e.g., "food", "transportation").
	Category string `json:"category" validate:"required,alphanumunderscore"`
	// Label is the human-readable name for the category (e.g., "Food & Dining", "Transportation").
	Label string `json:"label" validate:"required"`
	// PlannedAmount is the total amount planned to be spent for this category.
	PlannedAmount float64 `json:"plannedAmount" validate:"gte=0"`
	// ActualAmount is the total amount actually spent for this category.
	ActualAmount float64 `json:"actualAmount" validate:"gte=0"`
	// SpentPercentage is the percentage of the planned amount that has been spent (ActualAmount / PlannedAmount * 100).
	SpentPercentage float64 `json:"spentPercentage" validate:"gte=0"`
}

// PlannedVsActualRequest defines the query parameters for requesting
// the planned versus actual expenditure report.
// Month and Year can be optionally provided to specify the period.
// If not provided, the service typically defaults to the current month and year.
type PlannedVsActualRequest struct {
	// Month is the numerical representation of the month (1-12).
	// It's optional; if zero, the current month is usually assumed by the service.
	Month int `query:"month" validate:"omitempty,min=1,max=12"`
	// Year is the numerical representation of the year (e.g., 2024).
	// It's optional; if zero, the current year is usually assumed by the service.
	Year int `query:"year" validate:"omitempty,min=2020"`
}

// ExpensePlanningDoc represents the structure of an expense planning document
// as stored in Firestore. It outlines the planned budget for various categories
// for a specific user, month, and year.
type ExpensePlanningDoc struct {
	// ID is the document ID in Firestore, typically not part of the stored data itself but used for referencing.
	ID string `json:"-" firestore:"-"`
	// UserID is the identifier of the user to whom this expense plan belongs.
	UserID string `json:"userId" firestore:"userId"`
	// Month is the month (1-12) for which this plan is applicable.
	Month int `json:"month" firestore:"month"`
	// Year is the year for which this plan is applicable.
	Year int `json:"year" firestore:"year"`
	// Categories maps category keys (e.g., "food") to their planned amounts for the month.
	Categories map[string]float64 `json:"categories" firestore:"categories"`
	// CreatedAt is the timestamp when this planning document was created.
	CreatedAt time.Time `json:"created_at" firestore:"created_at"`
	// UpdatedAt is the timestamp when this planning document was last updated.
	UpdatedAt time.Time `json:"updated_at" firestore:"updated_at"`
}

// ExpenseDoc represents the structure of an individual expense transaction document
// as stored in Firestore. It details a single expense incurred by a user.
type ExpenseDoc struct {
	// ID is the document ID in Firestore.
	ID string `json:"-" firestore:"-"`
	// UserID is the identifier of the user who incurred the expense.
	UserID string `json:"userId" firestore:"user_id"`
	// Category is the key of the category this expense belongs to (e.g., "food").
	Category string `json:"category" firestore:"category"`
	// Amount is the monetary value of the expense.
	Amount float64 `json:"amount" firestore:"amount"`
	// PaymentDate is the date and time when the payment for this expense was made.
	PaymentDate time.Time `json:"payment_date" firestore:"payment_date"`
	// Status indicates the current state of the expense (e.g., "paid", "pending", "cancelled").
	Status string `json:"status" firestore:"status"`
	// Description provides additional details about the expense.
	Description string `json:"description" firestore:"description"`
	// CreatedAt is the timestamp when this expense document was created.
	CreatedAt time.Time `json:"created_at" firestore:"created_at"`
	// UpdatedAt is the timestamp when this expense document was last updated.
	UpdatedAt time.Time `json:"updated_at" firestore:"updated_at"`
}

// ExpenseCategoryDoc represents the structure for defining expense categories
// as stored in Firestore. This includes human-readable labels, icons, and other metadata
// associated with an expense category key.
type ExpenseCategoryDoc struct {
	// ID is the document ID in Firestore.
	ID string `json:"-" firestore:"-"`
	// Category is the unique key for the expense category (e.g., "food", "transportation").
	// This should match the keys used in ExpensePlanningDoc and ExpenseDoc.
	Category string `json:"category" firestore:"category"`
	// Label is the human-readable display name for the category (e.g., "Food & Dining", "Transportation").
	Label string `json:"label" firestore:"label"`
	// Icon is an optional identifier for an icon associated with the category (e.g., "icon-food").
	Icon string `json:"icon,omitempty" firestore:"icon,omitempty"`
	// Color is an optional hex color code associated with the category (e.g., "#FF0000").
	Color string `json:"color,omitempty" firestore:"color,omitempty"`
	// Active indicates whether this category is currently active and should be used/displayed.
	Active bool `json:"active" firestore:"active"`
}
