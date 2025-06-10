package entity_finance

import (
	"strings"
	"testing"
	"time"
)

func TestIncomeRecord_Validate(t *testing.T) {
	validDate := time.Now().Format("2006-01-02")
	one := 1 // Helper for pointer to int

	tests := []struct {
		name       string
		record     func() *IncomeRecord
		wantErrMsg string
	}{
		{
			name: "valid record - minimal",
			record: func() *IncomeRecord {
				return NewIncomeRecord("salary", "bankacc001", 100.0, validDate, "user001")
			},
			wantErrMsg: "",
		},
		{
			name: "valid record - full",
			record: func() *IncomeRecord {
				desc := "Full details"
				obs := "Some observations"
				rc := 12
				return &IncomeRecord{
					Category:        "freelance",
					Description:     &desc,
					BankAccountID:   "bankacc002",
					Amount:          1500.50,
					ReceiptDate:     validDate,
					IsRecurring:     true,
					RecurrenceCount: &rc,
					Observations:    &obs,
					UserID:          "user002",
				}
			},
			wantErrMsg: "",
		},
		{
			name: "missing category",
			record: func() *IncomeRecord {
				return NewIncomeRecord("", "bankacc001", 100.0, validDate, "user001")
			},
			wantErrMsg: "category is required",
		},
		{
			name: "invalid category",
			record: func() *IncomeRecord {
				return NewIncomeRecord("invalid_category_type", "bankacc001", 100.0, validDate, "user001")
			},
			wantErrMsg: "invalid category value",
		},
		{
			name: "description too long",
			record: func() *IncomeRecord {
				desc := strings.Repeat("a", 201)
				ir := NewIncomeRecord("salary", "bankacc001", 100.0, validDate, "user001")
				ir.Description = &desc
				return ir
			},
			wantErrMsg: "description must not exceed 200 characters",
		},
		{
			name: "missing bankAccountId",
			record: func() *IncomeRecord {
				return NewIncomeRecord("salary", "", 100.0, validDate, "user001")
			},
			wantErrMsg: "bankAccountId is required",
		},
		{
			name: "amount zero",
			record: func() *IncomeRecord {
				return NewIncomeRecord("salary", "bankacc001", 0.0, validDate, "user001")
			},
			wantErrMsg: "amount must be greater than 0",
		},
		{
			name: "amount negative",
			record: func() *IncomeRecord {
				return NewIncomeRecord("salary", "bankacc001", -10.0, validDate, "user001")
			},
			wantErrMsg: "amount must be greater than 0",
		},
		{
			name: "invalid receiptDate format",
			record: func() *IncomeRecord {
				return NewIncomeRecord("salary", "bankacc001", 100.0, "01-01-2023", "user001")
			},
			wantErrMsg: "receiptDate must be in YYYY-MM-DD format",
		},
		{
			name: "invalid receiptDate value",
			record: func() *IncomeRecord {
				return NewIncomeRecord("salary", "bankacc001", 100.0, "2023-13-01", "user001")
			},
			wantErrMsg: "invalid receiptDate: parsing time", // Error message contains details
		},
		{
			name: "isRecurring true but recurrenceCount nil",
			record: func() *IncomeRecord {
				ir := NewIncomeRecord("salary", "bankacc001", 100.0, validDate, "user001")
				ir.IsRecurring = true
				ir.RecurrenceCount = nil
				return ir
			},
			wantErrMsg: "recurrenceCount is required if isRecurring is true",
		},
		{
			name: "isRecurring true but recurrenceCount zero",
			record: func() *IncomeRecord {
				rc := 0
				ir := NewIncomeRecord("salary", "bankacc001", 100.0, validDate, "user001")
				ir.IsRecurring = true
				ir.RecurrenceCount = &rc
				return ir
			},
			wantErrMsg: "recurrenceCount must be at least 1 if isRecurring is true",
		},
		{
			name: "isRecurring true, recurrenceCount valid (1)",
			record: func() *IncomeRecord {
				ir := NewIncomeRecord("salary", "bankacc001", 100.0, validDate, "user001")
				ir.IsRecurring = true
				ir.RecurrenceCount = &one
				return ir
			},
			wantErrMsg: "",
		},
		{
			name: "observations too long",
			record: func() *IncomeRecord {
				obs := strings.Repeat("b", 501)
				ir := NewIncomeRecord("salary", "bankacc001", 100.0, validDate, "user001")
				ir.Observations = &obs
				return ir
			},
			wantErrMsg: "observations must not exceed 500 characters",
		},
		{
			name: "missing userID",
			record: func() *IncomeRecord {
				return NewIncomeRecord("salary", "bankacc001", 100.0, validDate, "")
			},
			wantErrMsg: "userID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ir := tt.record()
			err := ir.Validate()
			if tt.wantErrMsg == "" {
				if err != nil {
					t.Errorf("IncomeRecord.Validate() error = %v, wantErr %v", err, tt.wantErrMsg)
				}
			} else {
				if err == nil {
					t.Errorf("IncomeRecord.Validate() error = nil, wantErrMsg %s", tt.wantErrMsg)
				} else if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("IncomeRecord.Validate() error = %q, wantErrMsg %q", err.Error(), tt.wantErrMsg)
				}
			}
		})
	}
}

func TestGetIncomeRecordsQueryParameters_Validate(t *testing.T) {
	validStartDate := "2023-01-01"
	invalidDate := "01-01-2023"
	validEndDate := "2023-12-31"
	validSortKey := "amount"
	invalidSortKey := "name"
	validSortDir := "asc"
	invalidSortDir := "ascending"

	tests := []struct {
		name       string
		params     func() *GetIncomeRecordsQueryParameters
		wantErrMsg string
	}{
		{
			name: "valid params - empty",
			params: func() *GetIncomeRecordsQueryParameters {
				return &GetIncomeRecordsQueryParameters{}
			},
			wantErrMsg: "",
		},
		{
			name: "valid params - full",
			params: func() *GetIncomeRecordsQueryParameters {
				desc := "search term"
				return &GetIncomeRecordsQueryParameters{
					UserID:        "user123",
					Description:   &desc,
					StartDate:     &validStartDate,
					EndDate:       &validEndDate,
					SortKey:       &validSortKey,
					SortDirection: &validSortDir,
				}
			},
			wantErrMsg: "",
		},
		{
			name: "invalid startDate format",
			params: func() *GetIncomeRecordsQueryParameters {
				return &GetIncomeRecordsQueryParameters{StartDate: &invalidDate}
			},
			wantErrMsg: "invalid startDate format",
		},
		{
			name: "invalid endDate format",
			params: func() *GetIncomeRecordsQueryParameters {
				return &GetIncomeRecordsQueryParameters{EndDate: &invalidDate}
			},
			wantErrMsg: "invalid endDate format",
		},
		{
			name: "invalid sortKey",
			params: func() *GetIncomeRecordsQueryParameters {
				return &GetIncomeRecordsQueryParameters{SortKey: &invalidSortKey}
			},
			wantErrMsg: "invalid sortKey value",
		},
		{
			name: "invalid sortDirection",
			params: func() *GetIncomeRecordsQueryParameters {
				return &GetIncomeRecordsQueryParameters{SortDirection: &invalidSortDir}
			},
			wantErrMsg: "invalid sortDirection value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.params()
			err := p.Validate()
			if tt.wantErrMsg == "" {
				if err != nil {
					t.Errorf("GetIncomeRecordsQueryParameters.Validate() error = %v, wantErr %v", err, tt.wantErrMsg)
				}
			} else {
				if err == nil {
					t.Errorf("GetIncomeRecordsQueryParameters.Validate() error = nil, wantErrMsg %s", tt.wantErrMsg)
				} else if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("GetIncomeRecordsQueryParameters.Validate() error = %q, wantErrMsg %q", err.Error(), tt.wantErrMsg)
				}
			}
		})
	}
}

func TestIsValidIncomeCategory(t *testing.T) {
	tests := []struct {
		name     string
		category string
		want     bool
	}{
		{"valid salary", "salary", true},
		{"valid freelance", "freelance", true},
		{"valid rent_received", "rent_received", true},
		{"valid investment_income", "investment_income", true},
		{"valid bonus_plr", "bonus_plr", true},
		{"valid gift_donation", "gift_donation", true},
		{"valid reimbursement", "reimbursement", true},
		{"valid other", "other", true},
		{"invalid empty", "", false},
		{"invalid custom", "custom_cat", false},
		{"case sensitive", "Salary", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidIncomeCategory(tt.category); got != tt.want {
				t.Errorf("isValidIncomeCategory(%q) = %v, want %v", tt.category, got, tt.want)
			}
		})
	}
}
