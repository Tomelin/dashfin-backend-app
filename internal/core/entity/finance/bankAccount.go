package entity_finance

import (
	"errors"
	"strings"
)

type BankAccount struct {
	ID   string `json:"id,omitempty" bson:"_id,omitempty"`
	Name string `json:"name" bson:"name" validate:"required,min=3,max=100"`
	Code string `json:"code" bson:"code" validate:"required"`
	//BankCode       string  `json:"bankCode" bson:"bankCode" validate:"required"`
	//CustomBankName string  `json:"customBankName,omitempty" bson:"customBankName,omitempty" validate:"max=100"`
	//Description    string  `json:"description,omitempty" bson:"description,omitempty" validate:"max=150"`
	//Agency         string  `json:"agency" bson:"agency" validate:"required,min=3,max=10"`
	//AccountNumber  string  `json:"accountNumber" bson:"accountNumber" validate:"required,min=3,max=20"`
	//MonthlyFee     float64 `json:"monthlyFee" bson:"monthlyFee" validate:"min=0"`
}

func (ba *BankAccount) Validate() error {
	if err := ba.validateRequired(); err != nil {
		return err
	}

	return nil
}

func (ba *BankAccount) validateRequired() error {
	if strings.TrimSpace(ba.Code) == "" {
		return errors.New("code is required")
	}

	if strings.TrimSpace(ba.Name) == "" {
		return errors.New("name is required")
	}

	return nil
}
