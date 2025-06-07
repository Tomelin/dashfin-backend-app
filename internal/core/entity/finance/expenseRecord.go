package entity_finance

import (
	"context"
	"errors"
	"strings"
	"time"
)

// ExpenseRecordRepositoryInterface defines the repository operations for ExpenseRecord.
type ExpenseRecordRepositoryInterface interface {
	CreateExpenseRecord(ctx context.Context, data *ExpenseRecord) (*ExpenseRecord, error)
	GetExpenseRecordByID(ctx context.Context, id string) (*ExpenseRecord, error)
	GetExpenseRecords(ctx context.Context) ([]ExpenseRecord, error)
	GetExpenseRecordsByFilter(ctx context.Context, filter map[string]interface{}) ([]ExpenseRecord, error)
	UpdateExpenseRecord(ctx context.Context, id string, data *ExpenseRecord) (*ExpenseRecord, error)
	DeleteExpenseRecord(ctx context.Context, id string) error
}

// ExpenseRecordServiceInterface defines the service operations for ExpenseRecord.
type ExpenseRecordServiceInterface interface {
	CreateExpenseRecord(ctx context.Context, data *ExpenseRecord) (*ExpenseRecord, error)
	GetExpenseRecordByID(ctx context.Context, id string) (*ExpenseRecord, error)
	GetExpenseRecords(ctx context.Context) ([]ExpenseRecord, error)
	GetExpenseRecordsByFilter(ctx context.Context, filter map[string]interface{}) ([]ExpenseRecord, error)
	UpdateExpenseRecord(ctx context.Context, id string, data *ExpenseRecord) (*ExpenseRecord, error)
	DeleteExpenseRecord(ctx context.Context, id string) error
}

// ExpenseRecord defines the structure for an expense record.
type ExpenseRecord struct {
	ID              string    `json:"id" bson:"_id,omitempty"` // Auto-generated
	Category        string    `json:"category" bson:"category"`
	Subcategory     string    `json:"subcategory,omitempty" bson:"subcategory,omitempty"`
	DueDate         string    `json:"dueDate" bson:"dueDate"` // ISO 8601 (YYYY-MM-DD)
	PaymentDate     *string   `json:"paymentDate,omitempty" bson:"paymentDate,omitempty"` // ISO 8601 (YYYY-MM-DD)
	Amount          float64   `json:"amount" bson:"amount"`
	BankPaidFrom    *string   `json:"bankPaidFrom,omitempty" bson:"bankPaidFrom,omitempty"`
	CustomBankName  *string   `json:"customBankName,omitempty" bson:"customBankName,omitempty"`
	Description     *string   `json:"description,omitempty" bson:"description,omitempty"`
	IsRecurring     bool      `json:"isRecurring" bson:"isRecurring"`
	RecurrenceCount *int      `json:"recurrenceCount,omitempty" bson:"recurrenceCount,omitempty"`
	CreatedAt       time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt       time.Time `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	UserID          string    `json:"userId,omitempty" bson:"userId,omitempty"` // To associate with a user
}

// NewExpenseRecord creates a new ExpenseRecord with default values.
// Required fields (category, dueDate, amount) must be set separately.
func NewExpenseRecord(category, dueDate string, amount float64, userID string) *ExpenseRecord {
	return &ExpenseRecord{
		Category:    category,
		DueDate:     dueDate,
		Amount:      amount,
		IsRecurring: false,
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Validate checks the ExpenseRecord fields for correctness.
func (er *ExpenseRecord) Validate() error {
	if strings.TrimSpace(er.Category) == "" {
		return errors.New("category is required")
	}
	if len(er.Category) > 100 {
		return errors.New("category must not exceed 100 characters")
	}

	if er.Subcategory != "" && len(er.Subcategory) > 100 {
		return errors.New("subcategory must not exceed 100 characters")
	}

	if _, err := time.Parse("2006-01-02", er.DueDate); err != nil {
		return errors.New("dueDate must be in YYYY-MM-DD format")
	}

	if er.PaymentDate != nil && *er.PaymentDate != "" {
		if _, err := time.Parse("2006-01-02", *er.PaymentDate); err != nil {
			return errors.New("paymentDate must be in YYYY-MM-DD format if provided")
		}
	}

	if er.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}

	if er.PaymentDate != nil && *er.PaymentDate != "" {
		if er.BankPaidFrom == nil || strings.TrimSpace(*er.BankPaidFrom) == "" {
			return errors.New("bankPaidFrom is required if paymentDate is filled")
		}
	}

	if er.BankPaidFrom != nil && *er.BankPaidFrom == "other" {
		if er.CustomBankName == nil || strings.TrimSpace(*er.CustomBankName) == "" {
			return errors.New("customBankName is required when bankPaidFrom is 'other'")
		}
	}
    if er.CustomBankName != nil && len(*er.CustomBankName) > 100 {
        return errors.New("customBankName must not exceed 100 characters")
    }


	if er.Description != nil && len(*er.Description) > 200 {
		return errors.New("description must not exceed 200 characters")
	}

	if er.IsRecurring {
		if er.RecurrenceCount == nil || *er.RecurrenceCount < 1 {
			return errors.New("recurrenceCount must be at least 1 if isRecurring is true")
		}
	} else {
		if er.RecurrenceCount != nil {
			// Field should not be present if not recurring, or be explicitly nil/0
			// Depending on how you want to enforce this.
			// For now, let's say it's an error if it's set when not recurring.
			// return errors.New("recurrenceCount should only be set if isRecurring is true")
		}
	}
	if strings.TrimSpace(er.UserID) == "" {
		return errors.New("userID is required")
	}

	return nil
}
