package repository_finance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	entity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	"github.com/Tomelin/dashfin-backend-app/internal/core/repository"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

type CreditCardRepository struct {
	DB         database.FirebaseDBInterface
	collection string
}

func InitializeCreditCardRepository(db database.FirebaseDBInterface) (entity.CreditCardRepositoryInterface, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	return &CreditCardRepository{
		DB:         db,
		collection: "credit-card",
	}, nil
}

func (r *CreditCardRepository) CreateCreditCard(ctx context.Context, data *entity.CreditCard) (*entity.CreditCardRequest, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}

	toMap, _ := utils.StructToMap(data)

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, err
	}

	if collection == nil || *collection == "" {
		return nil, fmt.Errorf("%s collection is empty", r.collection)
	}

	doc, err := r.DB.Create(ctx, toMap, *collection)
	if err != nil {
		return nil, err
	}

	var response entity.CreditCardRequest
	err = json.Unmarshal(doc, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (r *CreditCardRepository) GetCreditCardByID(ctx context.Context, id *string) (*entity.CreditCardRequest, error) {
	if id == nil || *id == "" {
		return nil, errors.New("id is empty")
	}

	filters := map[string]interface{}{
		"id": *id,
	}

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, err
	}

	if collection == nil || *collection == "" {
		return nil, fmt.Errorf("%s collection is empty", r.collection)
	}

	result, err := r.DB.GetByFilter(ctx, filters, *collection)
	if err != nil {
		return nil, err
	}

	var CreditCards []entity.CreditCardRequest
	if err := json.Unmarshal(result, &CreditCards); err != nil {
		return nil, err
	}

	if len(CreditCards) == 0 {
		return nil, errors.New("bank account not found")
	}

	return &CreditCards[0], nil
}

func (r *CreditCardRepository) GetCreditCards(ctx context.Context) ([]entity.CreditCardRequest, error) {

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, err
	}

	result, err := r.DB.Get(ctx, *collection)
	if err != nil {
		return nil, err
	}

	var CreditCards []entity.CreditCardRequest
	if err := json.Unmarshal(result, &CreditCards); err != nil {
		return nil, err
	}

	return CreditCards, nil
}

func (r *CreditCardRepository) GetByFilter(ctx context.Context, data map[string]interface{}) ([]entity.CreditCardRequest, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, err
	}

	if collection == nil || *collection == "" {
		return nil, fmt.Errorf("%s collection is empty", r.collection)
	}

	result, err := r.DB.GetByFilter(ctx, data, *collection)
	if err != nil {
		return nil, err
	}

	var CreditCards []entity.CreditCardRequest
	if err := json.Unmarshal(result, &CreditCards); err != nil {
		return nil, err
	}

	return CreditCards, nil
}

func (r *CreditCardRepository) UpdateCreditCard(ctx context.Context, data *entity.CreditCardRequest) (*entity.CreditCardRequest, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}

	// Note: You'll need to add an ID field to the CreditCard entity for proper updates
	// For now, assuming there's a way to identify the document
	if data.ID == "" {
		return nil, errors.New("id is empty")
	}

	toMap, _ := utils.StructToMap(data)

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, err
	}
	if collection == nil || *collection == "" {
		return nil, fmt.Errorf("%s collection is empty", r.collection)
	}

	err = r.DB.Update(ctx, data.ID, toMap, *collection)
	if err != nil {
		return nil, fmt.Errorf("failed to update bank account: %w", err)
	}

	return data, nil
}

func (r *CreditCardRepository) DeleteCreditCard(ctx context.Context, id *string) error {
	if id == nil || *id == "" {
		return errors.New("id is empty")
	}

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return err
	}

	if collection == nil || *collection == "" {
		return fmt.Errorf("%s collection is empty", r.collection)
	}

	return r.DB.Delete(ctx, *id, *collection)
}
