package profile

import (
	"context"
	"errors"

	entity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
	repository "github.com/Tomelin/dashfin-backend-app/internal/core/repository/profile"
)

type ProfileGoalsServiceInterface interface {
	UpdateProfileGoals(ctx context.Context, userId *string, data *entity.ProfileGoals) (*entity.ProfileGoals, error)
	GetProfileGoals(ctx context.Context, userID *string) (entity.ProfileGoals, error)
}

type ProfileGoalsService struct {
	Repo    repository.ProfileRepositoryInterface
	Profile ProfilePersonServiceInterface
}

func InicializeProfileGoalsService(repo repository.ProfileRepositoryInterface, profile ProfilePersonServiceInterface) (ProfileGoalsServiceInterface, error) {

	if repo == nil {
		return nil, errors.New("repo is nil")
	}

	return &ProfileGoalsService{
		Repo:    repo,
		Profile: profile,
	}, nil
}

func (p *ProfileGoalsService) UpdateProfileGoals(ctx context.Context, userId *string, data *entity.ProfileGoals) (*entity.ProfileGoals, error) {
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
	profile.Goals = *data

	resultProfile, err := p.Repo.UpdateProfile(ctx, &profile)
	if err != nil {
		return nil, err
	}

	return &resultProfile.Goals, err
}

func (p *ProfileGoalsService) GetProfileGoals(ctx context.Context, userID *string) (entity.ProfileGoals, error) {
	if userID == nil {
		return entity.ProfileGoals{}, errors.New("user id is nil")
	}

	query := map[string]interface{}{
		"userProviderID": *userID,
	}
	profiles, err := p.Repo.GetByFilter(ctx, query)
	if err != nil {
		return entity.ProfileGoals{}, err
	}

	if profiles == nil {
		return entity.ProfileGoals{}, errors.New("user not found")
	}

	if len(profiles) == 0 {
		return entity.ProfileGoals{}, errors.New("user not found")
	}

	return profiles[0].Goals, nil
}
