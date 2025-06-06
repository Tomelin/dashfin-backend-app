package repository

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	entity_profile "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

type ProfileRepositoryInterface interface {
	CreateProfile(ctx context.Context, data *entity_profile.Profile) (*entity_profile.Profile, error)
	GetProfileByID(ctx context.Context, id string) (*entity_profile.Profile, error)
	GetProfile(ctx context.Context) ([]entity_profile.Profile, error)
	GetByFilter(ctx context.Context, data map[string]interface{}) ([]entity_profile.Profile, error)
	UpdateProfile(ctx context.Context, data *entity_profile.Profile) (*entity_profile.Profile, error)
}

type ProfileRepository struct {
	DB         database.FirebaseDBInterface
	collection string
}

func InicializeProfileRepository(db database.FirebaseDBInterface) (ProfileRepositoryInterface, error) {

	if db == nil {
		return nil, errors.New("db is nil")
	}

	return &ProfileRepository{
		DB:         db,
		collection: "profile",
	}, nil
}

func (r *ProfileRepository) CreateProfile(ctx context.Context, data *entity_profile.Profile) (*entity_profile.Profile, error) {

	if data == nil {
		return nil, errors.New("data is nil")
	}

	result, err := r.DB.Create(ctx, data, r.collection)
	if err != nil {
		return nil, err
	}

	var profile entity_profile.Profile
	err = json.Unmarshal(result, &profile)
	if err != nil {
		return nil, err
	}

	return &profile, err
}

func (r *ProfileRepository) UpdateProfile(ctx context.Context, data *entity_profile.Profile) (*entity_profile.Profile, error) {

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
	log.Println("Repo update err", result, err)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		log.Println("Repo update eln", result, len(result))
		return nil, errors.New("error get user after update")
	}
	log.Println("Repo update", result[0])
	return &result[0], err
}

func (r *ProfileRepository) GetProfileByID(ctx context.Context, id string) (*entity_profile.Profile, error) {

	if id == "" {
		return nil, errors.New("id is empty")
	}

	query := map[string]interface{}{
		"UserID": id,
	}

	results, err := r.DB.GetByFilter(ctx, query, r.collection)
	if err != nil {
		return nil, err
	}

	if len(results) > 0 {
		var profile []entity_profile.Profile
		err = json.Unmarshal(results, &profile)
		if err != nil {
			return nil, err
		}
		return &profile[0], nil
	}

	return nil, errors.New("user not found")
}

func (r *ProfileRepository) GetByFilter(ctx context.Context, data map[string]interface{}) ([]entity_profile.Profile, error) {

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

func (r *ProfileRepository) GetProfile(ctx context.Context) ([]entity_profile.Profile, error) {

	results, err := r.DB.Get(ctx, r.collection)
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
