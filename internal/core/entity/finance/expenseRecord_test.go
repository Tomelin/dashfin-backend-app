package entity_finance

import (
	"testing"
	"time"
)

func TestExpenseRecord_Validate(t *testing.T) {
	validRecord := func() *ExpenseRecord {
		return &ExpenseRecord{
			Category:    "Food",
			Subcategory: "Groceries",
			DueDate:     time.Now().Format("2006-01-02"),
			Amount:      50.00,
			UserID:      "user123",
			IsRecurring: false,
		}
	}

	tests := []struct {
		name    string
		er      *ExpenseRecord
		wantErr bool
		errText string
	}{
		{
			name:    "Valid record",
			er:      validRecord(),
			wantErr: false,
		},
		{
			name: "Missing Category",
			er: func() *ExpenseRecord {
				r := validRecord()
				r.Category = ""
				return r
			}(),
			wantErr: true,
			errText: "category is required",
		},
		{
			name: "Invalid DueDate format",
			er: func() *ExpenseRecord {
				r := validRecord()
				r.DueDate = "01-02-2023" // Invalid format
				return r
			}(),
			wantErr: true,
			errText: "dueDate must be in YYYY-MM-DD format",
		},
		{
			name: "Amount zero",
			er: func() *ExpenseRecord {
				r := validRecord()
				r.Amount = 0
				return r
			}(),
			wantErr: true,
			errText: "amount must be greater than 0",
		},
		{
			name: "Missing UserID",
			er: func() *ExpenseRecord {
				r := validRecord()
				r.UserID = ""
				return r
			}(),
			wantErr: true,
			errText: "userID is required",
		},
		{
			name: "Recurring but no RecurrenceCount",
			er: func() *ExpenseRecord {
				r := validRecord()
				r.IsRecurring = true
				r.RecurrenceCount = 0
				return r
			}(),
			wantErr: true,
			errText: "recurrenceCount must be at least 1 if isRecurring is true",
		},
		{
			name: "Recurring with RecurrenceCount zero",
			er: func() *ExpenseRecord {
				r := validRecord()
				r.IsRecurring = true
				zero := 0
				r.RecurrenceCount = zero
				return r
			}(),
			wantErr: true,
			errText: "recurrenceCount must be at least 1 if isRecurring is true",
		},
		{
			name: "Valid recurring record",
			er: func() *ExpenseRecord {
				r := validRecord()
				r.IsRecurring = true
				count := 12
				r.RecurrenceCount = count
				return r
			}(),
			wantErr: false,
		},
		{
			name: "BankPaidFrom 'other' but no CustomBankName",
			er: func() *ExpenseRecord {
				r := validRecord()
				paymentDate := time.Now().Format("2006-01-02")
				r.PaymentDate = &paymentDate
				otherBank := "other"
				r.BankPaidFrom = &otherBank
				r.CustomBankName = nil
				return r
			}(),
			wantErr: true,
			errText: "customBankName is required when bankPaidFrom is 'other'",
		},
		{
			name: "PaymentDate filled but BankPaidFrom missing",
			er: func() *ExpenseRecord {
				r := validRecord()
				paymentDate := time.Now().Format("2006-01-02")
				r.PaymentDate = &paymentDate
				r.BankPaidFrom = nil
				return r
			}(),
			wantErr: true,
			errText: "bankPaidFrom is required if paymentDate is filled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.er.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpenseRecord.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errText {
				t.Errorf("ExpenseRecord.Validate() error text = %q, want %q", err.Error(), tt.errText)
			}
		})
	}
}
