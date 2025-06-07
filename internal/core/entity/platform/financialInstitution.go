package entity_platform

import (
	"context"
	"fmt"
)

type FinancialInstitutionInterface interface {
	GetFinancialInstitutionByID(ctx context.Context, id *string) (*FinancialInstitution, error)
	GetAllFinancialInstitutions(ctx context.Context) ([]FinancialInstitution, error)
}

const (
	ErrInvalidID = "invalid financial institution ID"
)

type FinancialInstitution struct {
	Code     string `json:"code"  validate:"required"`
	Name     string `json:"name" validate:"required"`
	FullName string `json:"fullName,omitempty"`
}

func (f *FinancialInstitution) Validate() error {

	if f.Code == "" {
		return fmt.Errorf("code is required")
	}

	if f.Name == "" {
		return fmt.Errorf("name is required")
	}

	return nil
}
