package dto

import (
	"errors"
	"regexp"
	"strings"
	"time"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

type IncomeRecordDTO struct {
	ID               string  `json:"id" bson:"_id,omitempty"`
	Category         string  `json:"category"`
	Description      string  `json:"description,omitempty"`
	BankAccountID    string  `json:"bankAccountId"`
	Amount           float64 `json:"amount"`
	ReceiptDate      string  `json:"receiptDate"` // ISO 8601 (YYYY-MM-DD)
	IsRecurring      bool    `json:"isRecurring"`
	RecurrenceCount  int     `json:"recurrenceCount,omitempty" ` // Pointer to allow null
	RecurrenceNumber int     `json:"recurrenceNumber,omitempty"` // Pointer to allow null
	Observations     string  `json:"observations,omitempty"`
	UserID           string  `json:"userId,omitempty"` // To associate with a user
}

// Validate checks the IncomeRecord fields for correctness.
func (ir *IncomeRecordDTO) Validate() error {
	if strings.TrimSpace(ir.Category) == "" {
		return errors.New("category is required")
	}

	if len(ir.Description) > 200 {
		return errors.New("description must not exceed 200 characters")
	}

	if ir.Category == "" {
		return errors.New("category is required")
	}

	if strings.TrimSpace(ir.BankAccountID) == "" {
		return errors.New("bankAccountId is required")
	}

	if ir.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}

	// Validate ReceiptDate format (YYYY-MM-DD)
	if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, ir.ReceiptDate); !matched {
		return errors.New("receiptDate must be in YYYY-MM-DD format")
	}
	parsedReceiptDate, err := time.Parse("2006-01-02", ir.ReceiptDate)
	if err != nil {
		return errors.New("invalid receiptDate: " + err.Error())
	}
	ir.ReceiptDate = parsedReceiptDate.Format("2006-01-02") // Ensure consistent format

	if ir.IsRecurring {
		if ir.RecurrenceCount < 1 {
			return errors.New("recurrenceCount must be at least 1 if isRecurring is true")
		}

	} else {
		// If not recurring, RecurrenceCount should ideally be nil or ignored.
		// Depending on API design, you might want to enforce ir.RecurrenceCount == nil here.
	}

	if len(ir.Observations) > 500 {
		return errors.New("observations must not exceed 500 characters")
	}

	if strings.TrimSpace(ir.UserID) == "" {
		return errors.New("userID is required")
	}

	return nil
}

func (ir *IncomeRecordDTO) ToEntity() (*entity_finance.IncomeRecord, error) {
	if err := ir.Validate(); err != nil {
		return nil, err
	}

	receiptDate, err := utils.ConvertISO8601ToTime(ir.ReceiptDate)
	if err != nil {
		return nil, err
	}

	income := &entity_finance.IncomeRecord{
		ID:               ir.ID,
		Category:         ir.Category,
		Description:      ir.Description,
		BankAccountID:    ir.BankAccountID,
		Amount:           ir.Amount,
		ReceiptDate:      receiptDate,
		IsRecurring:      ir.IsRecurring,
		RecurrenceCount:  ir.RecurrenceCount,
		RecurrenceNumber: ir.RecurrenceNumber,
		Observations:     ir.Observations,
		UserID:           ir.UserID,
	}

	if err := income.Validate(); err != nil {
		return nil, err
	}

	return income, nil
}

func (ir *IncomeRecordDTO) FromEntity(income *entity_finance.IncomeRecord) {

	ir.ID = income.ID
	ir.Category = income.Category
	ir.Description = income.Description
	ir.BankAccountID = income.BankAccountID
	ir.Amount = income.Amount
	ir.ReceiptDate = utils.ConvertTimeToDateFormat(income.ReceiptDate)
	ir.IsRecurring = income.IsRecurring
	ir.RecurrenceCount = income.RecurrenceCount
	ir.RecurrenceNumber = income.RecurrenceNumber
	ir.Observations = income.Observations
	ir.UserID = income.UserID
}
