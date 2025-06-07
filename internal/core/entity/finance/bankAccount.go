package entity_finance

import (
	"errors"
	"strings"
)

type BankAccount struct {
	ID             string  `json:"id,omitempty" bson:"_id,omitempty"`
	BankCode       string  `json:"bankCode" bson:"bankCode" validate:"required"`
	CustomBankName string  `json:"customBankName,omitempty" bson:"customBankName,omitempty" validate:"max=100"`
	Description    string  `json:"description,omitempty" bson:"description,omitempty" validate:"max=150"`
	Agency         string  `json:"agency" bson:"agency" validate:"required,min=3,max=10"`
	AccountNumber  string  `json:"accountNumber" bson:"accountNumber" validate:"required,min=3,max=20"`
	MonthlyFee     float64 `json:"monthlyFee" bson:"monthlyFee" validate:"min=0"`
}

func (ba *BankAccount) Validate() error {
	if err := ba.validateRequired(); err != nil {
		return err
	}

	if err := ba.validateBankCode(); err != nil {
		return err
	}

	if err := ba.validateCustomBankName(); err != nil {
		return err
	}

	if err := ba.validateDescription(); err != nil {
		return err
	}

	if err := ba.validateAgency(); err != nil {
		return err
	}

	if err := ba.validateAccountNumber(); err != nil {
		return err
	}

	if err := ba.validateMonthlyFee(); err != nil {
		return err
	}

	return nil
}

func (ba *BankAccount) validateRequired() error {
	if strings.TrimSpace(ba.BankCode) == "" {
		return errors.New("bankCode is required")
	}

	if strings.TrimSpace(ba.Agency) == "" {
		return errors.New("agency is required")
	}

	if strings.TrimSpace(ba.AccountNumber) == "" {
		return errors.New("accountNumber is required")
	}

	return nil
}

func (ba *BankAccount) validateBankCode() error {
	bankCode := strings.TrimSpace(ba.BankCode)
	if bankCode == "" {
		return errors.New("bankCode cannot be empty")
	}

	return nil
}

func (ba *BankAccount) validateCustomBankName() error {
	if ba.BankCode == "other" && strings.TrimSpace(ba.CustomBankName) == "" {
		return errors.New("customBankName is required when bankCode is 'other'")
	}

	if len(ba.CustomBankName) > 100 {
		return errors.New("customBankName must not exceed 100 characters")
	}

	return nil
}

func (ba *BankAccount) validateDescription() error {
	if len(ba.Description) > 150 {
		return errors.New("description must not exceed 150 characters")
	}

	return nil
}

func (ba *BankAccount) validateAgency() error {
	agency := strings.TrimSpace(ba.Agency)
	if len(agency) < 3 {
		return errors.New("agency must have at least 3 characters")
	}

	if len(agency) > 10 {
		return errors.New("agency must not exceed 10 characters")
	}

	return nil
}

func (ba *BankAccount) validateAccountNumber() error {
	accountNumber := strings.TrimSpace(ba.AccountNumber)
	if len(accountNumber) < 3 {
		return errors.New("accountNumber must have at least 3 characters")
	}

	if len(accountNumber) > 20 {
		return errors.New("accountNumber must not exceed 20 characters")
	}

	return nil
}

func (ba *BankAccount) validateMonthlyFee() error {
	if ba.MonthlyFee < 0 {
		return errors.New("monthlyFee must be greater than or equal to 0")
	}

	return nil
}

func (ba *BankAccount) SetDefaults() {
	if ba.MonthlyFee == 0 {
		ba.MonthlyFee = 0.0
	}

	ba.BankCode = strings.TrimSpace(ba.BankCode)
	ba.CustomBankName = strings.TrimSpace(ba.CustomBankName)
	ba.Description = strings.TrimSpace(ba.Description)
	ba.Agency = strings.TrimSpace(ba.Agency)
	ba.AccountNumber = strings.TrimSpace(ba.AccountNumber)
}

func (ba *BankAccount) IsValid() bool {
	return ba.Validate() == nil
}
