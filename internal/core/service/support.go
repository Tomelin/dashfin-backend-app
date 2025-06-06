package service

import (
	"context"
	"errors"

	"github.com/Tomelin/dashfin-backend-app/internal/core/entity"
	"github.com/Tomelin/dashfin-backend-app/internal/core/repository"
)

type SupportServiceInterface interface {
	Create(ctx context.Context, data *entity.Support) (*entity.SupportResponse, error)
}

type SupportService struct {
	repo       repository.SupportRepositoryInterface
	collection string
}

func InicializeSupportService(repo repository.SupportRepositoryInterface) (SupportServiceInterface, error) {

	if repo == nil {
		return nil, errors.New("db is nil")
	}

	return &SupportService{
		repo:       repo,
		collection: "support",
	}, nil
}

func (s *SupportService) Create(ctx context.Context, data *entity.Support) (*entity.SupportResponse, error) {

	if data == nil {
		return nil, errors.New("data is nil")
	}

	if err := data.Validate(); err != nil {
		return nil, err
	}

	result, err := s.repo.Create(ctx, data)
	if err != nil {
		return nil, err
	}

	return result, nil
}
