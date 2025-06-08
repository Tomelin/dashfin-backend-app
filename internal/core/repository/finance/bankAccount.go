package repository_finance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	entity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

type BankAccountRepository struct {
	DB database.FirebaseDBInterface
}

func InitializeBankAccountRepository(db database.FirebaseDBInterface) (entity.BankAccountRepositoryInterface, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	return &BankAccountRepository{
		DB: db,
	}, nil
}

func (r *BankAccountRepository) CreateBankAccount(ctx context.Context, data *entity.BankAccount) (*entity.BankAccountRequest, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}

	toMap, _ := utils.StructToMap(data)

	doc, err := r.DB.Create(ctx, toMap, "bankAccounts")
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

	filters := map[string]interface{}{
		"id": *id,
	}

	result, err := r.DB.GetByFilter(ctx, filters, "bankAccounts")
	if err != nil {
		return nil, err
	}

	var bankAccounts []entity.BankAccountRequest
	if err := json.Unmarshal(result, &bankAccounts); err != nil {
		return nil, err
	}

	if len(bankAccounts) == 0 {
		return nil, errors.New("bank account not found")
	}

	return &bankAccounts[0], nil
}

func (r *BankAccountRepository) GetBankAccounts(ctx context.Context) ([]entity.BankAccountRequest, error) {
	result, err := r.DB.Get(ctx, "bankAccounts")
	if err != nil {
		return nil, err
	}

	log.Println(string(result))
	var bankAccounts []entity.BankAccountRequest
	if err := json.Unmarshal(result, &bankAccounts); err != nil {
		return nil, err
	}

	log.Println("accounts", bankAccounts)

	return bankAccounts, nil
}

func (r *BankAccountRepository) GetByFilter(ctx context.Context, data map[string]interface{}) ([]entity.BankAccountRequest, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}

	result, err := r.DB.GetByFilter(ctx, data, "bankAccounts")
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

	log.Println(">>>>> repository <<<<<")
	log.Println("ID", data.ID)
	log.Println("object", data)
	// Note: You'll need to add an ID field to the BankAccount entity for proper updates
	// For now, assuming there's a way to identify the document
	if data.ID == "" {
		return nil, errors.New("id is empty")
	}

	toMap, _ := utils.StructToMap(data)

	err := r.DB.Update(ctx, data.ID, toMap, "bankAccounts")
	if err != nil {
		return nil, fmt.Errorf("failed to update bank account: %w", err)
	}

	return data, nil
}

func (r *BankAccountRepository) DeleteBankAccount(ctx context.Context, id *string) error {
	if id == nil || *id == "" {
		return errors.New("id is empty")
	}

	return r.DB.Delete(ctx, *id, "bankAccounts")
}
