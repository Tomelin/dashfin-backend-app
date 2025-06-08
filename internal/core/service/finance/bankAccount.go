package service_finance

import (
	"context"
	"errors"
	"fmt"

	entity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
)

type BankAccountService struct {
	Repo entity.BankAccountRepositoryInterface
}

func InitializeBankAccountService(repo entity.BankAccountRepositoryInterface) (entity.BankAccountServiceInterface, error) {
	if repo == nil {
		return nil, errors.New("repo is nil")
	}

	return &BankAccountService{
		Repo: repo,
	}, nil
}

func (s *BankAccountService) CreateBankAccount(ctx context.Context, bankAccount *entity.BankAccount) (*entity.BankAccountRequest, error) {
	if bankAccount == nil {
		return nil, errors.New("bankAccount is nil")
	}

	if err := bankAccount.Validate(); err != nil {
		return nil, err
	}

	query := map[string]interface{}{
		"bankCode":      bankAccount.BankCode,
		"agency":        bankAccount.Agency,
		"accountNumber": bankAccount.AccountNumber,
	}

	results, err := s.GetByFilter(ctx, query)
	if err != nil && err.Error() != "bank account not found" {
		return nil, err
	}

	if len(results) > 0 {
		return nil, errors.New("bank account already exists")
	}

	result, err := s.Repo.CreateBankAccount(ctx, bankAccount)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *BankAccountService) GetBankAccountByID(ctx context.Context, id *string) (*entity.BankAccountRequest, error) {
	if id == nil || *id == "" {
		return nil, errors.New("id is empty")
	}

	result, err := s.Repo.GetBankAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, errors.New("bank account not found")
	}

	return result, nil
}

func (s *BankAccountService) GetBankAccounts(ctx context.Context) ([]entity.BankAccountRequest, error) {
	result, err := s.Repo.GetBankAccounts(ctx)
	if err != nil {
		return nil, err
	}

	if result == nil || len(result) == 0 {
		return nil, errors.New("bank accounts not found")
	}

	return result, nil
}

func (s *BankAccountService) GetByFilter(ctx context.Context, data map[string]interface{}) ([]entity.BankAccountRequest, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}

	result, err := s.Repo.GetByFilter(ctx, data)
	if err != nil {
		return nil, err
	}

	if result == nil || len(result) == 0 {
		return nil, errors.New("bank account not found")
	}

	return result, nil
}

func (s *BankAccountService) UpdateBankAccount(ctx context.Context, data *entity.BankAccountRequest) (*entity.BankAccountRequest, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}

	if err := data.Validate(); err != nil {
		return nil, fmt.Errorf("invalid data: %w", err)
	}

	return s.Repo.UpdateBankAccount(ctx, data)
}

func (s *BankAccountService) DeleteBankAccount(ctx context.Context, id *string) error {
	if id == nil || *id == "" {
		return errors.New("id is empty")
	}

	// _, err := s.GetBankAccountByID(ctx, id)
	// if err != nil {
	// 	return err
	// }

	return s.Repo.DeleteBankAccount(ctx, id)
}
