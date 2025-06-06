package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Tomelin/dashfin-backend-app/internal/core/entity"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

type SupportRepositoryInterface interface {
	Create(ctx context.Context, data *entity.Support) (*entity.SupportResponse, error)
}

type SupportRepository struct {
	DB         database.FirebaseDBInterface
	collection string
}

func InicializeSupportRepository(db database.FirebaseDBInterface) (SupportRepositoryInterface, error) {

	if db == nil {
		return nil, errors.New("db is nil")
	}

	return &SupportRepository{
		DB:         db,
		collection: "support",
	}, nil
}

func (s *SupportRepository) Create(ctx context.Context, data *entity.Support) (*entity.SupportResponse, error) {

	toMap, _ := utils.StructToMap(data)

	if toMap == nil {
		return nil, errors.New("data is nil")
	}

	result, err := s.DB.Create(ctx, toMap, s.collection)
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

	response := &entity.SupportResponse{
		ID:      id,
		Support: *data,
	}

	return response, nil
}
