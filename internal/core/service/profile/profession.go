package service

import (
	"context"
	"errors"

	entity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
	repository "github.com/Tomelin/dashfin-backend-app/internal/core/repository/profile"
)

type ProfileProfessionServiceInterface interface {
	UpdateProfileProfession(ctx context.Context, userId *string, data *entity.ProfileProfession) (*entity.ProfileProfession, error)
	// GetProfileProfessionByID(ctx context.Context, id *string) (*entity.Profile, error)
	// GetProfileProfession(ctx context.Context) ([]entity.Profile, error)
	// GetProfileProfessionByFilter(ctx context.Context, data map[string]interface{}) ([]entity.Profile, error)
	// UpdateProfileProfession(ctx context.Context, data *entity.Profile) (*entity.Profile, error)
}

type ProfileProfessionService struct {
	Repo    repository.ProfileRepositoryInterface
	Profile ProfilePersonServiceInterface
}

func InicializeProfileProfessionService(repo repository.ProfileRepositoryInterface, profile ProfilePersonServiceInterface) (ProfileProfessionServiceInterface, error) {

	if repo == nil {
		return nil, errors.New("repo is nil")
	}

	return &ProfileProfessionService{
		Repo:    repo,
		Profile: profile,
	}, nil
}

func (p *ProfileProfessionService) UpdateProfileProfession(ctx context.Context, userId *string, data *entity.ProfileProfession) (*entity.ProfileProfession, error) {
	if data == nil {
		return nil, errors.New("profession is nil")
	}

	if userId == nil {
		return nil, errors.New("user id is nil")
	}

	query := map[string]interface{}{
		"userProviderID": *userId,
	}

	profiles, err := p.Profile.GetByFilter(ctx, query)
	if err != nil {
		return nil, err
	}

	if profiles == nil {
		return nil, errors.New("user not found")
	}

	if len(profiles) == 0 {
		return nil, errors.New("user not found")
	}

	profile := profiles[0]
	profile.Profession = *data

	resultProfile, err := p.Repo.UpdateProfile(ctx, &profile)
	if err != nil {
		return nil, err
	}

	return &resultProfile.Profession, err
}

func (p *ProfileProfessionService) GetProfileProfessionByID(ctx context.Context, id *string) (*entity.Profile, error) {
	return nil, nil
}

func (p *ProfileProfessionService) GetProfileProfession(ctx context.Context) ([]entity.Profile, error) {
	return nil, nil
}

func (p *ProfileProfessionService) GetProfileProfessionByFilter(ctx context.Context, data map[string]interface{}) ([]entity.Profile, error) {
	return nil, nil
}

// func (p *ProfileProfessionService) UpdateProfileProfession(ctx context.Context, data *entity.Profile) (*entity.Profile, error) {
// 	return nil, nil
// }
