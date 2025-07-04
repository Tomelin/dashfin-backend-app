package dto

import (
	"errors"
	"regexp"
	"strings"
	"time"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

type ExpenseRecordDTO struct {
	ID               string    `json:"id"`
	Amount           float64   `json:"amount"`
	Category         string    `json:"category"`
	DueDate          string    `json:"dueDate"`
	IsRecurring      bool      `json:"isRecurring"`
	UserID           string    `json:"userId,omitempty"`
	BankPaidFrom     string    `json:"bankPaidFrom,omitempty"`
	CustomBankName   string    `json:"customBankName,omitempty"`
	Description      string    `json:"description,omitempty"`
	PaymentDate      string    `json:"paymentDate,omitempty"`
	RecurrenceCount  int       `json:"recurrenceCount,omitempty"`
	RecurrenceNumber int       `json:"recurrenceNumber,omitempty"`
	Subcategory      string    `json:"subcategory,omitempty"`
	CreatedAt        time.Time `json:"createdAt,omitempty"`
	UpdatedAt        time.Time `json:"updatedAt,omitempty"`
}

func (er *ExpenseRecordDTO) Validate() error {
	if er.Category == "" {
		return errors.New("category is required")
	}

	if er.DueDate == "" {
		return errors.New("dueDate is required")
	}

	if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, er.DueDate); !matched {
		return errors.New("DueDate must be in YYYY-MM-DD format")
	}
	parsedDueDate, err := time.Parse("2006-01-02", er.DueDate)
	if err != nil {
		return errors.New("invalid dueDate: " + err.Error())
	}
	er.DueDate = parsedDueDate.Format("2006-01-02") // Ensure consistent format

	if _, err := time.Parse("2006-01-02", er.DueDate); err != nil {
		return errors.New("dueDate must be in YYYY-MM-DD format")
	}

	if er.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}

	if er.PaymentDate != "" || er.BankPaidFrom != "" {

		if er.PaymentDate == "" {
			return errors.New("paymentDate is required if bankPaidFrom is filled")
		}

		if er.BankPaidFrom == "other" && er.CustomBankName == "" {
			return errors.New("customBankName is required when bankPaidFrom is 'other'")
		}
		if _, err := time.Parse("2006-01-02", er.PaymentDate); err != nil {
			return errors.New("paymentDate must be in YYYY-MM-DD format")
		}

		if er.Description != "" && len(er.Description) > 200 {
			return errors.New("description must not exceed 200 characters")
		}

		if strings.TrimSpace(er.UserID) == "" {
			return errors.New("userID is required")
		}

		if er.IsRecurring && er.RecurrenceCount <= 0 {
			return errors.New("recurrenceCount must be greater than 0 when isRecurring is true")
		}
	}

	return nil
}

func (er *ExpenseRecordDTO) ToEntity() (*entity_finance.ExpenseRecord, error) {

	if er.Validate() != nil {
		return nil, er.Validate()
	}

	dueDate, err := utils.ConvertISO8601ToTime(er.DueDate)
	if err != nil {
		return nil, err
	}

	var paymentDate time.Time
	if er.PaymentDate != "" {
		paymentDate, err = utils.ConvertISO8601ToTime(er.PaymentDate)
		if err != nil {
			return nil, err
		}
	}

	expense := &entity_finance.ExpenseRecord{
		ID:               er.ID,
		Category:         er.Category,
		Subcategory:      er.Subcategory,
		DueDate:          dueDate,
		PaymentDate:      paymentDate,
		Amount:           er.Amount,
		BankPaidFrom:     er.BankPaidFrom,
		CustomBankName:   er.CustomBankName,
		Description:      er.Description,
		IsRecurring:      er.IsRecurring,
		RecurrenceNumber: er.RecurrenceNumber,
		RecurrenceCount:  er.RecurrenceCount,
		CreatedAt:        er.CreatedAt,
		UpdatedAt:        er.UpdatedAt,
		UserID:           er.UserID,
	}

	if expense.Validate() != nil {
		return nil, expense.Validate()
	}

	return expense, nil
}

func (er *ExpenseRecordDTO) FromEntity(expense *entity_finance.ExpenseRecord) {

	er.ID = expense.ID
	er.Category = expense.Category
	er.Subcategory = expense.Subcategory
	er.DueDate = utils.ConvertTimeToDateFormat(expense.DueDate)
	if !expense.PaymentDate.IsZero() {
		er.PaymentDate = utils.ConvertTimeToDateFormat(expense.PaymentDate)
	} else {
		er.PaymentDate = ""
	}
	er.Amount = expense.Amount
	er.BankPaidFrom = expense.BankPaidFrom
	er.CustomBankName = expense.CustomBankName
	er.Description = expense.Description
	er.IsRecurring = expense.IsRecurring
	er.RecurrenceNumber = expense.RecurrenceNumber
	er.RecurrenceCount = expense.RecurrenceCount
	er.CreatedAt = expense.CreatedAt
	er.UpdatedAt = expense.UpdatedAt
	er.UserID = expense.UserID

}
