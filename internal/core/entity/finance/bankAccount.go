package entity_finance

import (
	"context"
	"errors"
	"strings"
)

type BankAccountRepositoryInterface interface {
	CreateBankAccount(ctx context.Context, data *BankAccount) (*BankAccountRequest, error)
	GetBankAccountByID(ctx context.Context, id *string) (*BankAccountRequest, error)
	GetBankAccounts(ctx context.Context) ([]BankAccountRequest, error)
	GetByFilter(ctx context.Context, data map[string]interface{}) ([]BankAccountRequest, error)
	UpdateBankAccount(ctx context.Context, data *BankAccountRequest) (*BankAccountRequest, error)
	DeleteBankAccount(ctx context.Context, id *string) error
}

type BankAccountServiceInterface interface {
	CreateBankAccount(ctx context.Context, data *BankAccount) (*BankAccountRequest, error)
	GetBankAccountByID(ctx context.Context, id *string) (*BankAccountRequest, error)
	GetBankAccounts(ctx context.Context) ([]BankAccountRequest, error)
	GetByFilter(ctx context.Context, data map[string]interface{}) ([]BankAccountRequest, error)
	UpdateBankAccount(ctx context.Context, data *BankAccountRequest) (*BankAccountRequest, error)
	DeleteBankAccount(ctx context.Context, id *string) error
}

type BankAccount struct {
	BankCode       string  `json:"bankCode" bson:"bankCode"`
	CustomBankName string  `json:"customBankName,omitempty" bson:"customBankName,omitempty"`
	Description    string  `json:"description,omitempty" bson:"description,omitempty"`
	Agency         string  `json:"agency" bson:"agency"`
	AccountNumber  string  `json:"accountNumber" bson:"accountNumber"`
	MonthlyFee     float64 `json:"monthlyFee" bson:"monthlyFee"`
}

type BankAccountRequest struct {
	ID string `json:"id"`
	BankAccount
}
type BankAccountResponse BankAccountRequest

func NewBankAccount(bankCode, agency, accountNumber string) *BankAccount {
	return &BankAccount{
		BankCode:      bankCode,
		Agency:        agency,
		AccountNumber: accountNumber,
		MonthlyFee:    0.0,
	}
}

func (ba *BankAccount) Validate() error {
	if strings.TrimSpace(ba.BankCode) == "" {
		return errors.New("bankCode is required")
	}

	if ba.BankCode == "other" && strings.TrimSpace(ba.CustomBankName) == "" {
		return errors.New("customBankName is required when bankCode is 'other'")
	}

	if len(ba.CustomBankName) > 100 {
		return errors.New("customBankName must not exceed 100 characters")
	}

	if len(ba.Description) > 150 {
		return errors.New("description must not exceed 150 characters")
	}

	agency := strings.TrimSpace(ba.Agency)
	if agency == "" {
		return errors.New("agency is required")
	}
	if len(agency) < 3 || len(agency) > 10 {
		return errors.New("agency must be between 3 and 10 characters")
	}

	accountNumber := strings.TrimSpace(ba.AccountNumber)
	if accountNumber == "" {
		return errors.New("accountNumber is required")
	}
	if len(accountNumber) < 3 || len(accountNumber) > 20 {
		return errors.New("accountNumber must be between 3 and 20 characters")
	}

	if ba.MonthlyFee < 0 {
		return errors.New("monthlyFee must be greater than or equal to 0")
	}

	return nil
}
