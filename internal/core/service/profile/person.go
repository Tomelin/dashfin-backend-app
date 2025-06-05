package service

import (
	"context"
	"errors"
	"log"

	entity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
	repository "github.com/Tomelin/dashfin-backend-app/internal/core/repository/profile"
)

type ProfileServiceInterface interface {
	GetProfile(id string) (interface{}, error)
	CreateProfile(ctx context.Context, profile *entity.Profile) (interface{}, error)
}

type ProfileService struct {
	Repo repository.ProfileRepositoryInterface
}

func InicializeProfileService(repo repository.ProfileRepositoryInterface) (ProfileServiceInterface, error) {

	if repo == nil {
		return nil, errors.New("repo is nil")
	}

	return &ProfileService{
		Repo: repo,
	}, nil
}

func (s *ProfileService) GetProfile(id string) (interface{}, error) {
	return s.Repo.GetProfile(id)
}

func (s *ProfileService) CreateProfile(ctx context.Context, profile *entity.Profile) (interface{}, error) {
	result, err := s.Repo.CreateProfile(ctx, *profile)
	log.Println("Service result")
	log.Println(result, err)
	if err != nil {
		return nil, err
	}

	return result, err
}
