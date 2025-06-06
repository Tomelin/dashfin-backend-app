package repository

import (
	"context"
	"encoding/json"
	"errors"

	entity_profile "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

type ProfileProfessionRepositoryInterface interface {

	// GetProfileByID(ctx context.Context, id *string) (*entity_profile.ProfileProfession, error)
	// GetProfileProfession(ctx context.Context) ([]entity_profile.ProfileProfession, error)
	GetByFilter(ctx context.Context, data map[string]interface{}) ([]entity_profile.Profile, error)
	UpdateProfileProfession(ctx context.Context, data *entity_profile.Profile) (*entity_profile.Profile, error)
}

type ProfileProfessionRepository struct {
	DB         database.FirebaseDBInterface
	collection string
}

func InicializeProfileProfessionRepository(db database.FirebaseDBInterface) (ProfileProfessionRepositoryInterface, error) {

	if db == nil {
		return nil, errors.New("db is nil")
	}

	return &ProfileProfessionRepository{
		DB:         db,
		collection: "profiles",
	}, nil
}

func (r *ProfileProfessionRepository) UpdateProfileProfession(ctx context.Context, data *entity_profile.Profile) (*entity_profile.Profile, error) {

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

// func (r *ProfileProfessionRepository) GetProfileByID(ctx context.Context, id *string) (*entity_profile.ProfileProfession, error) {
// 	if id == nil {
// 		return nil, errors.New("id is empty")
// 	}

// 	query := map[string]interface{}{
// 		"userProviderID": *id,
// 	}

// 	results, err := r.DB.GetByFilter(ctx, query, r.collection)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var profiles []entity_profile.ProfileProfession
// 	err = json.Unmarshal(results, &profiles)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if len(profiles) == 0 {
// 		return nil, errors.New("user not found")
// 	}

// 	return &profiles[0], nil
// }

func (r *ProfileProfessionRepository) GetByFilter(ctx context.Context, data map[string]interface{}) ([]entity_profile.Profile, error) {

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

// func (r *ProfileProfessionRepository) GetProfileProfession(ctx context.Context) ([]entity_profile.ProfileProfession, error) {

// 	results, err := r.DB.Get(ctx, r.collection)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var profile []entity_profile.ProfileProfession
// 	err = json.Unmarshal(results, &profile)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return profile, nil
// }
