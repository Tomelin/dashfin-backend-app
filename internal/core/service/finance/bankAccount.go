package service

import (
	"context"
	"errors"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	repository_finance "github.com/Tomelin/dashfin-backend-app/internal/core/repository/finance"
)

type BankAccountServiceInterface interface {
	CreateBankAccount(ctx context.Context, userID string, data *entity_finance.BankAccount) (*entity_finance.BankAccount, error)
	GetBankAccountByID(ctx context.Context, userID, id string) (*entity_finance.BankAccount, error)
	GetBankAccounts(ctx context.Context, userID string) ([]entity_finance.BankAccount, error)
	UpdateBankAccount(ctx context.Context, userID, id string, data *entity_finance.BankAccount) (*entity_finance.BankAccount, error)
	DeleteBankAccount(ctx context.Context, userID, id string) error
}

type BankAccountService struct {
	Repo repository_finance.BankAccountRepositoryInterface
}

func InitializeBankAccountService(repo repository_finance.BankAccountRepositoryInterface) (BankAccountServiceInterface, error) {
	if repo == nil {
		return nil, errors.New("repo is nil")
	}

	return &BankAccountService{
		Repo: repo,
	}, nil
}

func (s *BankAccountService) CreateBankAccount(ctx context.Context, userID string, bankAccount *entity_finance.BankAccount) (*entity_finance.BankAccount, error) {
	if bankAccount == nil {
		return nil, errors.New("bankAccount is nil")
	}

	if userID == "" {
		return nil, errors.New("userID is required")
	}

	if err := bankAccount.Validate(); err != nil {
		return nil, err
	}

	result, err := s.Repo.CreateBankAccount(ctx, userID, bankAccount)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *BankAccountService) GetBankAccountByID(ctx context.Context, userID, id string) (*entity_finance.BankAccount, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	if id == "" {
		return nil, errors.New("id is required")
	}

	result, err := s.Repo.GetBankAccountByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, errors.New("bank account not found")
	}

	return result, nil
}

func (s *BankAccountService) GetBankAccounts(ctx context.Context, userID string) ([]entity_finance.BankAccount, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	result, err := s.Repo.GetBankAccounts(ctx, userID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *BankAccountService) UpdateBankAccount(ctx context.Context, userID, id string, bankAccount *entity_finance.BankAccount) (*entity_finance.BankAccount, error) {
	if bankAccount == nil {
		return nil, errors.New("bankAccount is nil")
	}

	if userID == "" {
		return nil, errors.New("userID is required")
	}

	if id == "" {
		return nil, errors.New("id is required")
	}

	if err := bankAccount.Validate(); err != nil {
		return nil, err
	}

	result, err := s.Repo.UpdateBankAccount(ctx, userID, id, bankAccount)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *BankAccountService) DeleteBankAccount(ctx context.Context, userID, id string) error {
	if userID == "" {
		return errors.New("userID is required")
	}

	if id == "" {
		return errors.New("id is required")
	}

	return s.Repo.DeleteBankAccount(ctx, userID, id)
}
