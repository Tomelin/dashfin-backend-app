package repository_finance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	"github.com/Tomelin/dashfin-backend-app/internal/core/repository"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

// SpendingPlanRepository handles database operations for pendingPlan.
type SpendingPlanRepository struct {
	DB         database.FirebaseDBInterface
	collection string
}

// InitializeSpendingPlanRepository creates a new SpendingPlanRepository.
func InitializeSpendingPlanRepository(db database.FirebaseDBInterface) (entity_finance.SpendingPlanRepositoryInterface, error) {
	if db == nil {
		return nil, errors.New("database is nil for IncomeRecordRepository")
	}

	return &SpendingPlanRepository{
		DB:         db,
		collection: fmt.Sprintf("%s_spending_plan", dbPath),
	}, nil
}

func (r *SpendingPlanRepository) GetSpendingPlanByUserID(ctx context.Context, userID string) (*entity_finance.SpendingPlan, error) {
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}

	filters := map[string]interface{}{
		"id": userID,
	}

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		log.Println("err > ", err)
		return nil, err
	}

	docs, err := r.DB.Get(ctx, *collection)
	if err != nil {
		return nil, err
	}

	var records []entity_finance.SpendingPlan
	if err := json.Unmarshal(docs, &records); err != nil {
		return nil, err
	}

	var result *entity_finance.SpendingPlan
	for _, v := range records {
		if v.UserID == filters["id"] {
			result = &v
			break
		}
	}

	if result == nil {
		return nil, errors.New("spendingPlan not found")
	}

	return result, nil
}

func (r *SpendingPlanRepository) UpdateSpendingPlan(ctx context.Context, data *entity_finance.SpendingPlan) error {

	if data == nil {
		return errors.New("plan cannot be nil")
	}

	data.UpdatedAt = time.Now()

	toMap, _ := utils.StructToMap(data)

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return err
	}

	log.Println(" repo > ", data.UserID, toMap)
	err = r.DB.Update(ctx, data.UserID, toMap, *collection)
	if err != nil {
		return err
	}

	return nil
}

func (r *SpendingPlanRepository) CreateSpendingPlan(ctx context.Context, data *entity_finance.SpendingPlan) (*entity_finance.SpendingPlan, error) {

	if data == nil {
		return nil, errors.New("plan cannot be nil")
	}

	// Generate ID and set timestamps
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()

	toMap, _ := utils.StructToMap(data)

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, err
	}

	doc, err := r.DB.Create(ctx, toMap, *collection)
	if err != nil {
		return nil, err
	}

	var response entity_finance.SpendingPlan
	err = json.Unmarshal(doc, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
