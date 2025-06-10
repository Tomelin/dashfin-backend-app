package service_platform

import (
	"context"
	"errors"

	entity_platform "github.com/Tomelin/dashfin-backend-app/internal/core/entity/platform"
)

type FinancialInstitutionService struct {
	repo entity_platform.FinancialInstitutionInterface
}

func NewFinancialInstitutionService(repo entity_platform.FinancialInstitutionInterface) (entity_platform.FinancialInstitutionInterface, error) {
	if repo == nil {
		return nil, errors.New("repository cannot be nil")
	}

	return &FinancialInstitutionService{
		repo: repo,
	}, nil
}

func (fis *FinancialInstitutionService) GetFinancialInstitutionByID(ctx context.Context, id *string) (*entity_platform.FinancialInstitution, error) {
	if id == nil {
		return nil, errors.New(entity_platform.ErrInvalidID)

	}
	return fis.repo.GetFinancialInstitutionByID(ctx, id)
}

func (fis *FinancialInstitutionService) GetAllFinancialInstitutions(ctx context.Context) ([]entity_platform.FinancialInstitution, error) {
	return fis.repo.GetAllFinancialInstitutions(ctx)
}
