package repository

import (
	"context"
	"encoding/json"
	"errors"

	entity_profile "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

type ProfileGoalsRepositoryInterface interface {

	// GetProfileByID(ctx context.Context, id *string) (*entity_profile.ProfileGoals, error)
	// GetProfileGoals(ctx context.Context) ([]entity_profile.ProfileGoals, error)
	GetByFilter(ctx context.Context, data map[string]interface{}) ([]entity_profile.Profile, error)
	UpdateProfileGoals(ctx context.Context, data *entity_profile.Profile) (*entity_profile.Profile, error)
}

type ProfileGoalsRepository struct {
	DB         database.FirebaseDBInterface
	collection string
}

func InicializeProfileGoalsRepository(db database.FirebaseDBInterface) (ProfileGoalsRepositoryInterface, error) {

	if db == nil {
		return nil, errors.New("db is nil")
	}

	return &ProfileGoalsRepository{
		DB:         db,
		collection: "profiles",
	}, nil
}

func (r *ProfileGoalsRepository) UpdateProfileGoals(ctx context.Context, data *entity_profile.Profile) (*entity_profile.Profile, error) {

	if data == nil {
		return nil, errors.New("data is nil")
	}

	toMap, _ := utils.StructToMap(data)

	err := r.DB.Update(ctx, data.ID, toMap, r.collection)
	if err != nil {
		return nil, err
	}

	result, err := r.GetByFilter(ctx, map[string]interface{}{
		"userProviderID": data.UserProviderID,
	})
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, errors.New("error get user after update")
	}

	return &result[0], err
}

func (r *ProfileGoalsRepository) GetByFilter(ctx context.Context, data map[string]interface{}) ([]entity_profile.Profile, error) {

	if data == nil {
		return nil, errors.New("data is nil")
	}

	results, err := r.DB.GetByFilter(ctx, data, r.collection)
	if err != nil {
		return nil, err
	}

	var profile []entity_profile.Profile
	err = json.Unmarshal(results, &profile)
	if err != nil {
		return nil, err
	}
	return profile, nil
}
