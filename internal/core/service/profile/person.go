package service

import (
	"context"
	"errors"

	entity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
	repository "github.com/Tomelin/dashfin-backend-app/internal/core/repository/profile"
)

type ProfileServiceInterface interface {
	CreateProfile(ctx context.Context, data *entity.Profile) (*entity.Profile, error)
	GetProfileByID(ctx context.Context, id string) (*entity.Profile, error)
	GetProfile(ctx context.Context) ([]entity.Profile, error)
	GetByFilter(ctx context.Context, data map[string]interface{}) ([]entity.Profile, error)
	UpdateProfile(ctx context.Context, data *entity.Profile) (*entity.Profile, error)
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

func (s *ProfileService) CreateProfile(ctx context.Context, profile *entity.Profile) (*entity.Profile, error) {

	if profile == nil {
		return nil, errors.New("profile is nil")
	}

	query := map[string]interface{}{
		"Email": profile.Email,
	}

	results, err := s.GetByFilter(ctx, query)
	if err != nil {
		if err.Error() != "profile not found" {
			return nil, err
		}
	}

	if len(results) > 0 {
		return nil, errors.New("user already exists")
	}

	result, err := s.Repo.CreateProfile(ctx, profile)

	if err != nil {
		return nil, err
	}

	return result, err
}

func (s *ProfileService) GetProfileByID(ctx context.Context, id string) (*entity.Profile, error) {

	if id == "" {
		return nil, errors.New("id is empty")
	}

	result, err := s.Repo.GetProfileByID(ctx, id)

	return result, err
}

func (s *ProfileService) GetProfile(ctx context.Context) ([]entity.Profile, error) {

	result, err := s.Repo.GetProfile(ctx)
	if result == nil {
		return nil, errors.New("profile not found")
	}

	if len(result) == 0 {
		return nil, errors.New("profile not found")
	}

	return result, err

}

func (s *ProfileService) GetByFilter(ctx context.Context, data map[string]interface{}) ([]entity.Profile, error) {

	if data == nil {
		return nil, errors.New("data is nil")
	}

	result, err := s.Repo.GetByFilter(ctx, data)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, errors.New("profile not found 1")
	}

	if len(result) == 0 {
		return nil, errors.New("profile not found 2")
	}

	return result, err
}

func (s *ProfileService) UpdateProfile(ctx context.Context, data *entity.Profile) (*entity.Profile, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}

	// Buscar por UserProviderID (mais confiável que email)
	results, err := s.GetByFilter(ctx, map[string]interface{}{
		"userProviderID": data.UserProviderID, // campo correto
	})
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, errors.New("profile not found")
	}

	// Usar o ID do registro existente
	data.ID = results[0].ID

	return s.Repo.UpdateProfile(ctx, data)
}
