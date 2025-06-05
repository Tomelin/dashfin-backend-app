package repository

import (
	"context"
	"errors"
	"log"

	"github.com/Tomelin/dashfin-backend-app/pkg/database"
)

type ProfileRepositoryInterface interface {
	CreateProfile(ctx context.Context, data interface{}) (interface{}, error)
	GetProfile(id string) (interface{}, error)
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

func (r *ProfileRepository) CreateProfile(ctx context.Context, data interface{}) (interface{}, error) {

	result, err := r.DB.Create(ctx, data, r.collection)
	log.Println("Repository result")
	log.Println(result, err)
	if err != nil {
		return nil, err
	}
	return result, err
}

func (r *ProfileRepository) GetProfile(id string) (interface{}, error) {
	return nil, nil
}
