package service_finance

import (
	"context"
	"errors"
	"fmt"

	entity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
)

type CreditCardService struct {
	Repo entity.CreditCardRepositoryInterface
}

func InitializeCreditCardService(repo entity.CreditCardRepositoryInterface) (entity.CreditCardServiceInterface, error) {
	if repo == nil {
		return nil, errors.New("repo is nil")
	}

	return &CreditCardService{
		Repo: repo,
	}, nil
}

func (s *CreditCardService) CreateCreditCard(ctx context.Context, creditCard *entity.CreditCard) (*entity.CreditCardRequest, error) {
	if creditCard == nil {
		return nil, errors.New("creditCard is nil")
	}

	if err := creditCard.Validate(); err != nil {
		return nil, err
	}

	query := map[string]interface{}{
		"cardBrand":      creditCard.CardBrand,
		"lastFourDigits": creditCard.LastFourDigits,
	}

	results, err := s.GetByFilter(ctx, query)
	if err != nil && err.Error() != "credit card not found" {
		return nil, err
	}

	if len(results) > 0 {
		return nil, errors.New("credit card already exists")
	}

	result, err := s.Repo.CreateCreditCard(ctx, creditCard)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *CreditCardService) GetCreditCardByID(ctx context.Context, id *string) (*entity.CreditCardRequest, error) {
	if id == nil || *id == "" {
		return nil, errors.New("id is empty")
	}

	result, err := s.Repo.GetCreditCardByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, errors.New("credit card not found")
	}

	return result, nil
}

func (s *CreditCardService) GetCreditCards(ctx context.Context) ([]entity.CreditCardRequest, error) {
	result, err := s.Repo.GetCreditCards(ctx)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, errors.New("credit cards not found")
	}

	return result, nil
}

func (s *CreditCardService) GetByFilter(ctx context.Context, data map[string]interface{}) ([]entity.CreditCardRequest, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}

	result, err := s.Repo.GetByFilter(ctx, data)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, errors.New("credit card not found")
	}

	return result, nil
}

func (s *CreditCardService) UpdateCreditCard(ctx context.Context, data *entity.CreditCardRequest) (*entity.CreditCardRequest, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}

	if err := data.Validate(); err != nil {
		return nil, fmt.Errorf("invalid data: %w", err)
	}

	return s.Repo.UpdateCreditCard(ctx, data)
}

func (s *CreditCardService) DeleteCreditCard(ctx context.Context, id *string) error {
	if id == nil || *id == "" {
		return errors.New("id is empty")
	}

	return s.Repo.DeleteCreditCard(ctx, id)
}
