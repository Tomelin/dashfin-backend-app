package repository

import (
	"context"
	"encoding/json"
	"errors"

	entity_profile "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

type ProfileRepositoryInterface interface {
	CreateProfile(ctx context.Context, data *entity_profile.Profile) (*entity_profile.Profile, error)
	GetProfileByID(ctx context.Context, id *string) (*entity_profile.Profile, error)
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
		collection: "profiles",
	}, nil
}

func (r *ProfileRepository) CreateProfile(ctx context.Context, data *entity_profile.Profile) (*entity_profile.Profile, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}

	toMap, _ := utils.StructToMap(data)

	result, err := r.DB.Create(ctx, toMap, r.collection)
	if err != nil {
		return nil, err
	}

	// O Create retorna apenas o ID, não o objeto completo
	var docRef map[string]interface{}
	err = json.Unmarshal(result, &docRef)
	if err != nil {
		return nil, err
	}

	// Definir o ID no objeto original
	var id string
	_, ok := docRef["id"]
	if ok {
		id = docRef["id"].(string)
	} else {
		id = docRef["ID"].(string)
	}
	data.ID = id

	return data, nil
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
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, errors.New("error get user after update")
	}

	return &result[0], err
}

func (r *ProfileRepository) GetProfileByID(ctx context.Context, id *string) (*entity_profile.Profile, error) {
	if id == nil {
		return nil, errors.New("id is empty")
	}

	query := map[string]interface{}{
		"userProviderID": *id,
	}

	results, err := r.DB.GetByFilter(ctx, query, r.collection)
	if err != nil {
		return nil, err
	}

	var profiles []entity_profile.Profile
	err = json.Unmarshal(results, &profiles)
	if err != nil {
		return nil, err
	}

	if len(profiles) == 0 {
		return nil, errors.New("user not found")
	}

	return &profiles[0], nil
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
