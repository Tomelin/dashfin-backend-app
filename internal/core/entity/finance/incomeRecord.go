package entity_finance

import (
	"context"
	"errors"
	"strings"
	"time"

	entity_common "github.com/Tomelin/dashfin-backend-app/internal/core/entity/common"
)

// IncomeRecordRepositoryInterface defines the repository operations for IncomeRecord.
type IncomeRecordRepositoryInterface interface {
	CreateIncomeRecord(ctx context.Context, data *IncomeRecord) (*IncomeRecord, error)
	GetIncomeRecordByID(ctx context.Context, id string) (*IncomeRecord, error)
	GetIncomeRecords(ctx context.Context, params *GetIncomeRecordsQueryParameters) ([]IncomeRecord, error)
	UpdateIncomeRecord(ctx context.Context, id string, data *IncomeRecord) (*IncomeRecord, error)
	DeleteIncomeRecord(ctx context.Context, id string) error
}

// IncomeRecordServiceInterface defines the service operations for IncomeRecord.
type IncomeRecordServiceInterface interface {
	CreateIncomeRecord(ctx context.Context, data *IncomeRecord) (*IncomeRecord, error)
	GetIncomeRecordByID(ctx context.Context, id string) (*IncomeRecord, error)
	GetIncomeRecords(ctx context.Context, params *GetIncomeRecordsQueryParameters) ([]IncomeRecord, error)
	UpdateIncomeRecord(ctx context.Context, id string, data *IncomeRecord) (*IncomeRecord, error)
	DeleteIncomeRecord(ctx context.Context, id string) error
}

// IncomeRecord defines the structure for an income record.
type IncomeRecord struct {
	ID               string    `json:"id" bson:"_id,omitempty"` // Auto-generated
	Category         string    `json:"category" bson:"category"`
	Description      *string   `json:"description,omitempty" bson:"description,omitempty"`
	BankAccountID    string    `json:"bankAccountId" bson:"bankAccountId"`
	Amount           float64   `json:"amount" bson:"amount"`
	ReceiptDate      time.Time `json:"receiptDate" bson:"receiptDate"` // ISO 8601 (YYYY-MM-DD)
	IsRecurring      bool      `json:"isRecurring" bson:"isRecurring"`
	RecurrenceCount  *int      `json:"recurrenceCount,omitempty" bson:"recurrenceCount,omitempty"`   // Pointer to allow null
	RecurrenceNumber int       `json:"recurrenceNumber,omitempty" bson:"recurrenceNumber,omitempty"` // Pointer to allow null
	Observations     *string   `json:"observations,omitempty" bson:"observations,omitempty"`
	UserID           string    `json:"userId,omitempty" bson:"userId,omitempty"` // To associate with a user
	CreatedAt        time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt        time.Time `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	// TotalValue might not be needed if it's always equal to Amount for incomes
}

type IncomeRecordEvent struct {
	Action entity_common.ActionEvent `json:"action"` // "created", "updated", "deleted"
	Data   IncomeRecord              `json:"data"`
}

// IncomeCategory represents the allowed categories for income.
type IncomeCategory string

const (
	Salary           IncomeCategory = "salary"
	Freelance        IncomeCategory = "freelance"
	RentReceived     IncomeCategory = "rent_received"
	InvestmentIncome IncomeCategory = "investment_income"
	BonusPLR         IncomeCategory = "bonus_plr"
	GiftDonation     IncomeCategory = "gift_donation"
	Reimbursement    IncomeCategory = "reimbursement"
	Other            IncomeCategory = "other"
)

var validIncomeCategories = map[IncomeCategory]bool{
	Salary:           true,
	Freelance:        true,
	RentReceived:     true,
	InvestmentIncome: true,
	BonusPLR:         true,
	GiftDonation:     true,
	Reimbursement:    true,
	Other:            true,
}

// isValidIncomeCategory checks if the provided category is valid.
func isValidIncomeCategory(category string) bool {
	_, ok := validIncomeCategories[IncomeCategory(category)]
	return ok
}

// NewIncomeRecord creates a new IncomeRecord with default values.
// Required fields must be set separately.
func NewIncomeRecord(category, bankAccountID string, amount float64, receiptDate time.Time, userID string) *IncomeRecord {
	return &IncomeRecord{
		Category:      category,
		BankAccountID: bankAccountID,
		Amount:        amount,
		ReceiptDate:   receiptDate,
		IsRecurring:   false, // Default value
		UserID:        userID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// Validate checks the IncomeRecord fields for correctness.
func (ir *IncomeRecord) Validate() error {
	if strings.TrimSpace(ir.Category) == "" {
		return errors.New("category is required")
	}
	if !isValidIncomeCategory(ir.Category) {
		return errors.New("invalid category value")
	}

	if ir.Description != nil && len(*ir.Description) > 200 {
		return errors.New("description must not exceed 200 characters")
	}

	if strings.TrimSpace(ir.BankAccountID) == "" {
		return errors.New("bankAccountId is required")
	}

	if ir.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}

	// Validate ReceiptDate format (YYYY-MM-DD)
	ir.ReceiptDate, _ = time.Parse("2006-01-02", ir.ReceiptDate.Format("2006-01-02"))
	if ir.ReceiptDate.IsZero() {
		return errors.New("invalid receiptDate format, expected YYYY-MM-DD")
	}

	if ir.IsRecurring {
		if ir.RecurrenceCount == nil {
			return errors.New("recurrenceCount is required if isRecurring is true")
		}
		if *ir.RecurrenceCount < 1 {
			return errors.New("recurrenceCount must be at least 1 if isRecurring is true")
		}
	} else {
		// If not recurring, RecurrenceCount should ideally be nil or ignored.
		// Depending on API design, you might want to enforce ir.RecurrenceCount == nil here.
	}

	if ir.Observations != nil && len(*ir.Observations) > 500 {
		return errors.New("observations must not exceed 500 characters")
	}

	if strings.TrimSpace(ir.UserID) == "" {
		return errors.New("userID is required")
	}

	return nil
}

// GetIncomeRecordsQueryParameters defines the structure for filtering and sorting income records.
// This can be used by the service and repository layers.
type GetIncomeRecordsQueryParameters struct {
	UserID        string
	Description   *string `json:"description,omitempty"`   // Textual search
	StartDate     *string `json:"startDate,omitempty"`     // YYYY-MM-DD
	EndDate       *string `json:"endDate,omitempty"`       // YYYY-MM-DD
	SortKey       *string `json:"sortKey,omitempty"`       // "category", "amount", "receiptDate"
	SortDirection *string `json:"sortDirection,omitempty"` // "asc", "desc"
}

func (p *GetIncomeRecordsQueryParameters) Validate() error {
	if p.StartDate != nil {
		if _, err := time.Parse("2006-01-02", *p.StartDate); err != nil {
			return errors.New("invalid startDate format, expected YYYY-MM-DD")
		}
	}
	if p.EndDate != nil {
		if _, err := time.Parse("2006-01-02", *p.EndDate); err != nil {
			return errors.New("invalid endDate format, expected YYYY-MM-DD")
		}
	}
	if p.SortKey != nil {
		validKeys := map[string]bool{"category": true, "amount": true, "receiptDate": true}
		if !validKeys[*p.SortKey] {
			return errors.New("invalid sortKey value")
		}
	}
	if p.SortDirection != nil {
		validDirections := map[string]bool{"asc": true, "desc": true}
		if !validDirections[*p.SortDirection] {
			return errors.New("invalid sortDirection value, expected 'asc' or 'desc'")
		}
	}
	return nil
}
