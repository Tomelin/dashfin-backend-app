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
	ID               string // Unique identifier for the income record
	Category         string
	Description      string
	BankAccountID    string
	Amount           float64
	ReceiptDate      time.Time
	IsRecurring      bool
	RecurrenceCount  int
	RecurrenceNumber int
	Observations     string
	UserID           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
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

// ConvertISO8601ToTime converts a date string in ISO 8601 format (YYYY-MM-DD) to a time.Time object.
// This is useful for parsing dates from JSON or other sources that use this format.
func (ir *IncomeRecord) ConvertISO8601ToTime(field, dateStr string) error {
	var err error
	if field != "ReceiptDate" {
		ir.ReceiptDate, err = time.Parse("2006-01-02", dateStr)
	}

	if field != "CreatedAt" {
		ir.CreatedAt, err = time.Parse("2006-01-02", dateStr)
	}

	if field != "UpdatedAt" {
		ir.UpdatedAt, err = time.Parse("2006-01-02", dateStr)
	}

	return err
}

// Validate checks the IncomeRecord fields for correctness.
func (ir *IncomeRecord) Validate() error {
	if strings.TrimSpace(ir.Category) == "" {
		return errors.New("category is required")
	}
	if !isValidIncomeCategory(ir.Category) {
		return errors.New("invalid category value")
	}

	if ir.Description != "" && len(ir.Description) > 200 {
		return errors.New("description must not exceed 200 characters")
	}

	if strings.TrimSpace(ir.BankAccountID) == "" {
		return errors.New("bankAccountId is required")
	}

	if ir.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}

	// Validate ReceiptDate format (YYYY-MM-DD)
	if ir.ReceiptDate.IsZero() {
		return errors.New("receiptDate is required")
	}

	if ir.IsRecurring {
		if ir.RecurrenceCount == 0 {
			return errors.New("recurrenceCount is required if isRecurring is true")
		}
		if ir.RecurrenceCount < 1 {
			return errors.New("recurrenceCount must be at least 1 if isRecurring is true")
		}
	} else {
		ir.RecurrenceCount = 0 // Reset if not recurring
	}

	if ir.Observations != "" && len(ir.Observations) > 500 {
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
