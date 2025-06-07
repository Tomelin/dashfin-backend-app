package repository_platform

import (
	"context"
	"encoding/json"
	"errors"

	entity_platform "github.com/Tomelin/dashfin-backend-app/internal/core/entity/platform"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
)

type financialInstitutionRepository struct {
	DB         database.FirebaseDBInterface
	collection string
}

func NewFinancialInstitutionRepository(db database.FirebaseDBInterface) (entity_platform.FinancialInstitutionInterface, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	return &financialInstitutionRepository{
		DB:         db,
		collection: "platform_financial-institution",
	}, nil
}

func (r *financialInstitutionRepository) GetFinancialInstitutionByID(ctx context.Context, id *string) (*entity_platform.FinancialInstitution, error) {
	if id == nil {
		return nil, errors.New(entity_platform.ErrInvalidID)
	}

	query := map[string]interface{}{
		"id": *id,
	}

	results, err := r.DB.GetByFilter(ctx, query, r.collection)
	if err != nil {
		return nil, err
	}

	var institution entity_platform.FinancialInstitution
	err = json.Unmarshal(results, &institution)
	if err != nil {
		return nil, err
	}
	return &institution, nil
}

func (r *financialInstitutionRepository) GetAllFinancialInstitutions(ctx context.Context) ([]entity_platform.FinancialInstitution, error) {
	results, err := r.DB.Get(ctx, r.collection)
	if err != nil {
		return nil, err
	}

	var institutions []entity_platform.FinancialInstitution
	err = json.Unmarshal(results, &institutions)
	if err != nil {
		return nil, err
	}
	return institutions, nil
}
