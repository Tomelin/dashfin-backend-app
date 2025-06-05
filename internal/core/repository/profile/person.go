package repository

import (
	"context"
	"errors"

	entity_profile "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
)

type ProfileRepositoryInterface interface {
	CreateProfile(ctx context.Context, data *entity_profile.Profile) (interface{}, error)
	GetProfile(id string) (interface{}, error)
	GetByFilter(ctx context.Context, data map[string]interface{}) ([]interface{}, error)
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

func (r *ProfileRepository) CreateProfile(ctx context.Context, data *entity_profile.Profile) (interface{}, error) {

	if data == nil {
		return nil, errors.New("data is nil")
	}

	query := map[string]interface{}{
		"Email": data.Email,
	}

	results, err := r.GetByFilter(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(results) > 0 {
		return nil, errors.New("user already exists")
	}

	result, err := r.DB.Create(ctx, data, r.collection)
	if err != nil {
		return nil, err
	}
	return result, err
}

func (r *ProfileRepository) GetProfile(id string) (interface{}, error) {
	return nil, nil
}

func (r *ProfileRepository) GetByFilter(ctx context.Context, data map[string]interface{}) ([]interface{}, error) {

	if data == nil {
		return nil, errors.New("data is nil")
	}

	results, err := r.DB.GetByFilter(ctx, data, r.collection)
	if err != nil {
		return nil, err
	}

	return results, err
}
