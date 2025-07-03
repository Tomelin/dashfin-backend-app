package entity_finance

import (
	"context"
	"errors"
	"net/url"
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
	GetExpenseRecordsByDate(ctx context.Context, filter *ExpenseRecordQueryByDate) ([]ExpenseRecord, error)
	UpdateExpenseRecord(ctx context.Context, id string, data *ExpenseRecord) (*ExpenseRecord, error)
	DeleteExpenseRecord(ctx context.Context, id string) error
	CreateExpenseByNfceUrl(ctx context.Context, url *ExpenseByNfceUrl) (*ExpenseByNfceUrl, error)
}

// ExpenseRecord defines the structure for an expense record.
type ExpenseRecord struct {
	ID               string
	Category         string
	Subcategory      string
	DueDate          time.Time
	PaymentDate      time.Time
	Amount           float64
	BankPaidFrom     string
	CustomBankName   string
	Description      string
	IsRecurring      bool
	RecurrenceNumber int
	RecurrenceCount  int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	UserID           string
}

type ExpenseRecordQueryByDate struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

// NewExpenseRecord creates a new ExpenseRecord with default values.
// Required fields (category, dueDate, amount) must be set separately.
func NewExpenseRecord(category string, dueDate time.Time, amount float64, userID string) *ExpenseRecord {
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

type NfceUrl string

const (
	NfceUrlItems NfceUrl = "item-by-item"
	NfceUrlTotal NfceUrl = "total-value"
)

type ExpenseByNfceUrl struct {
	NfceUrl    string  `json:"nfceUrl" binding:"required"`
	UserID     string  `json:"userId"`
	ImportMode NfceUrl `json:"importMode" binding:"required"`
}

type NFCeItem struct {
	ItemDescription string  `json:"item_description"`
	ItemPrice       float64 `json:"item_price"`
}

type NFCe struct {
	Itens    []NFCeItem `json:"itens"`
	IDSeller string     `json:"id_seller"`
}

func (ex *ExpenseByNfceUrl) Validate() error {
	if strings.TrimSpace(ex.NfceUrl) == "" {
		return errors.New("nfceUrl is required")
	}

	if strings.TrimSpace(ex.UserID) == "" {
		return errors.New("userId is required")
	}

	if strings.TrimSpace(string(ex.ImportMode)) == "" {
		return errors.New("importMode is required")
	}

	if ex.ImportMode != NfceUrlItems && ex.ImportMode != NfceUrlTotal {
		return errors.New("importMode must be either 'item-by-item' or 'total-value'")
	}

	parsedURL, err := url.Parse(ex.NfceUrl)
	if err != nil {
		return errors.New("nfceUrl must be a valid URL")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("nfceUrl must use http or https protocol")
	}

	return nil
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

	if er.DueDate.IsZero() {
		return errors.New("dueDate is required")
	}

	if er.PaymentDate != (time.Time{}) || er.BankPaidFrom != "" {
		if er.PaymentDate.IsZero() {
			return errors.New("paymentDate is required if bankPaidFrom is filled")
		}
	}

	if er.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}

	if er.BankPaidFrom != "" && er.BankPaidFrom == "other" {
		if er.CustomBankName == "" || strings.TrimSpace(er.CustomBankName) == "" {
			return errors.New("customBankName is required when bankPaidFrom is 'other'")
		}
	}
	if er.CustomBankName != "" && len(er.CustomBankName) > 100 {
		return errors.New("customBankName must not exceed 100 characters")
	}

	if er.Description != "" && len(er.Description) > 200 {
		return errors.New("description must not exceed 200 characters")
	}

	if er.IsRecurring && er.RecurrenceCount < 0 {
		return errors.New("recurrenceCount must be greater than 0 when isRecurring is true")
	}

	if strings.TrimSpace(er.UserID) == "" {
		return errors.New("userID is required")
	}

	return nil
}
