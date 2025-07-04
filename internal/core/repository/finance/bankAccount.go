package repository_finance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	entity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	"github.com/Tomelin/dashfin-backend-app/internal/core/repository"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

type BankAccountRepository struct {
	DB         database.FirebaseDBInterface
	collection string
}

func InitializeBankAccountRepository(db database.FirebaseDBInterface) (entity.BankAccountRepositoryInterface, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	return &BankAccountRepository{
		DB:         db,
		collection: "bank-accounts",
	}, nil
}

func (r *BankAccountRepository) CreateBankAccount(ctx context.Context, data *entity.BankAccount) (*entity.BankAccountRequest, error) {
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

	var response entity.BankAccountRequest
	err = json.Unmarshal(doc, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (r *BankAccountRepository) GetBankAccountByID(ctx context.Context, id *string) (*entity.BankAccountRequest, error) {
	if id == nil || *id == "" {
		return nil, errors.New("id is empty")
	}

	conditional := []database.Conditional{
		{
			Field:  "id",
			Value:  *id,
			Filter: database.FilterEquals,
		},
	}

	if len(conditional) == 0 {
		return nil, errors.New("no conditions provided")
	}

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, err
	}

	if collection == nil || *collection == "" {
		return nil, fmt.Errorf("%s collection is empty", r.collection)
	}

	result, err := r.DB.GetByConditional(ctx, conditional, *collection)
	if err != nil {
		return nil, err
	}

	var bankAccounts []entity.BankAccountRequest
	if err := json.Unmarshal(result, &bankAccounts); err != nil {
		return nil, err
	}

	log.Println("\n GetBankAccountByID result:", bankAccounts, "collection:", *collection)

	if len(bankAccounts) == 0 {
		return nil, errors.New("bank account not found")
	}

	return &bankAccounts[0], nil
}

func (r *BankAccountRepository) GetBankAccounts(ctx context.Context) ([]entity.BankAccountRequest, error) {

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, err
	}

	result, err := r.DB.Get(ctx, *collection)
	if err != nil {
		return nil, err
	}

	var bankAccounts []entity.BankAccountRequest
	if err := json.Unmarshal(result, &bankAccounts); err != nil {
		return nil, err
	}

	return bankAccounts, nil
}

func (r *BankAccountRepository) GetByFilter(ctx context.Context, data map[string]interface{}) ([]entity.BankAccountRequest, error) {
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

	var bankAccounts []entity.BankAccountRequest
	if err := json.Unmarshal(result, &bankAccounts); err != nil {
		return nil, err
	}

	return bankAccounts, nil
}

func (r *BankAccountRepository) UpdateBankAccount(ctx context.Context, data *entity.BankAccountRequest) (*entity.BankAccountRequest, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}

	// Note: You'll need to add an ID field to the BankAccount entity for proper updates
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

func (r *BankAccountRepository) DeleteBankAccount(ctx context.Context, id *string) error {
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
