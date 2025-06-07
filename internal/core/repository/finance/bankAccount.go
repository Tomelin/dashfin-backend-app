package repository

import (
	"context"
	"encoding/json"
	"errors"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

type BankAccountRepositoryInterface interface {
	CreateBankAccount(ctx context.Context, userID string, data *entity_finance.BankAccount) (*entity_finance.BankAccount, error)
	GetBankAccountByID(ctx context.Context, userID, id string) (*entity_finance.BankAccount, error)
	GetBankAccounts(ctx context.Context, userID string) ([]entity_finance.BankAccount, error)
	UpdateBankAccount(ctx context.Context, userID, id string, data *entity_finance.BankAccount) (*entity_finance.BankAccount, error)
	DeleteBankAccount(ctx context.Context, userID, id string) error
}

type BankAccountRepository struct {
	DB         database.FirebaseDBInterface
	collection string
}

func InitializeBankAccountRepository(db database.FirebaseDBInterface) (BankAccountRepositoryInterface, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	return &BankAccountRepository{
		DB:         db,
		collection: "bank_accounts",
	}, nil
}

func (r *BankAccountRepository) CreateBankAccount(ctx context.Context, userID string, data *entity_finance.BankAccount) (*entity_finance.BankAccount, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}

	if userID == "" {
		return nil, errors.New("userID is required")
	}

	if err := data.Validate(); err != nil {
		return nil, err
	}

	toMap, _ := utils.StructToMap(data)
	toMap["userID"] = userID

	result, err := r.DB.Create(ctx, toMap, r.collection)
	if err != nil {
		return nil, err
	}

	var docRef map[string]interface{}
	err = json.Unmarshal(result, &docRef)
	if err != nil {
		return nil, err
	}

	var id string
	if docRef["id"] != nil {
		id = docRef["id"].(string)
	} else {
		id = docRef["ID"].(string)
	}
	data.ID = id

	return data, nil
}

func (r *BankAccountRepository) GetBankAccountByID(ctx context.Context, userID, id string) (*entity_finance.BankAccount, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	if id == "" {
		return nil, errors.New("id is required")
	}

	query := map[string]interface{}{
		"userID": userID,
		"id":     id,
	}

	results, err := r.DB.GetByFilter(ctx, query, r.collection)
	if err != nil {
		return nil, err
	}

	var bankAccounts []entity_finance.BankAccount
	err = json.Unmarshal(results, &bankAccounts)
	if err != nil {
		return nil, err
	}

	if len(bankAccounts) == 0 {
		return nil, errors.New("bank account not found")
	}

	return &bankAccounts[0], nil
}

func (r *BankAccountRepository) GetBankAccounts(ctx context.Context, userID string) ([]entity_finance.BankAccount, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	query := map[string]interface{}{
		"userID": userID,
	}

	results, err := r.DB.GetByFilter(ctx, query, r.collection)
	if err != nil {
		return nil, err
	}

	var bankAccounts []entity_finance.BankAccount
	err = json.Unmarshal(results, &bankAccounts)
	if err != nil {
		return nil, err
	}

	return bankAccounts, nil
}

func (r *BankAccountRepository) UpdateBankAccount(ctx context.Context, userID, id string, data *entity_finance.BankAccount) (*entity_finance.BankAccount, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}

	if userID == "" {
		return nil, errors.New("userID is required")
	}

	if id == "" {
		return nil, errors.New("id is required")
	}

	if err := data.Validate(); err != nil {
		return nil, err
	}

	existing, err := r.GetBankAccountByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	data.ID = existing.ID

	toMap, _ := utils.StructToMap(data)
	toMap["userID"] = userID

	err = r.DB.Update(ctx, data.ID, toMap, r.collection)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *BankAccountRepository) DeleteBankAccount(ctx context.Context, userID, id string) error {
	if userID == "" {
		return errors.New("userID is required")
	}

	if id == "" {
		return errors.New("id is required")
	}

	_, err := r.GetBankAccountByID(ctx, userID, id)
	if err != nil {
		return err
	}

	return r.DB.Delete(ctx, id, r.collection)
}
