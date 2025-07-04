package repository_finance

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	"github.com/Tomelin/dashfin-backend-app/internal/core/repository"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

// ExpenseRecordRepository handles database operations for ExpenseRecords.
type ExpenseRecordRepository struct {
	DB         database.FirebaseDBInterface
	collection string
}

// InitializeExpenseRecordRepository creates a new ExpenseRecordRepository.
func InitializeExpenseRecordRepository(db database.FirebaseDBInterface) (entity_finance.ExpenseRecordRepositoryInterface, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	return &ExpenseRecordRepository{
		DB:         db,
		collection: "expenses",
	}, nil
}

// CreateExpenseRecord adds a new expense record to the database.
func (r *ExpenseRecordRepository) CreateExpenseRecord(ctx context.Context, data *entity_finance.ExpenseRecord) (*entity_finance.ExpenseRecord, error) {
	if data == nil {
		return nil, errors.New("expense record data is nil")
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

	var response interface{}
	err = json.Unmarshal(doc, &response)
	if err != nil {
		return nil, err
	}

	responseEntity, err := r.convertToEntity(response)
	if err != nil {
		log.Println("[RESPONSE] Error converting response to entity:", err)
		return nil, err
	}

	return &responseEntity[0], nil
}

// GetExpenseRecordByID retrieves an expense record by its ID.
func (r *ExpenseRecordRepository) GetExpenseRecordByID(ctx context.Context, id string) (*entity_finance.ExpenseRecord, error) {
	if id == "" {
		return nil, errors.New("id is empty")
	}

	filters := map[string]interface{}{
		"id": id,
	}

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, err
	}

	docs, err := r.DB.Get(ctx, *collection)
	if err != nil {
		return nil, err
	}

	var response interface{}
	err = json.Unmarshal(docs, &response)
	if err != nil {
		return nil, err
	}

	responseEntity, err := r.convertToEntity(response)
	if err != nil {
		log.Println("[RESPONSE] Error converting response to entity:", err)
		return nil, err
	}

	var result *entity_finance.ExpenseRecord
	for _, v := range responseEntity {
		if v.ID == filters["id"] {
			result = &v
			break
		}
	}

	return result, nil
}

// GetExpenseRecords retrieves all expense records for the user in context (or all if no user context).
func (r *ExpenseRecordRepository) GetExpenseRecords(ctx context.Context) ([]entity_finance.ExpenseRecord, error) {

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, err
	}

	result, err := r.DB.Get(ctx, *collection)
	if err != nil {
		return nil, err
	}

	var response interface{}
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, err
	}

	responseEntity, err := r.convertToEntity(response)
	if err != nil {
		log.Println("[RESPONSE] Error converting response to entity:", err)
		return nil, err
	}

	return responseEntity, nil
}

// GetExpenseRecordsByFilter retrieves expense records based on a filter.
func (r *ExpenseRecordRepository) GetExpenseRecordsByFilter(ctx context.Context, filter map[string]interface{}) ([]entity_finance.ExpenseRecord, error) {
	if filter == nil {
		return nil, errors.New("filter data is nil")
	}

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, err
	}

	result, err := r.DB.GetByFilter(ctx, filter, *collection)
	if err != nil {
		return nil, err
	}

	var response interface{}
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, err
	}

	responseEntity, err := r.convertToEntity(response)
	if err != nil {
		log.Println("[RESPONSE] Error converting response to entity:", err)
		return nil, err
	}

	return responseEntity, nil
}

// UpdateExpenseRecord updates an existing expense record.
func (r *ExpenseRecordRepository) UpdateExpenseRecord(ctx context.Context, id string, data *entity_finance.ExpenseRecord) (*entity_finance.ExpenseRecord, error) {
	if id == "" {
		return nil, errors.New("id is empty for update")
	}
	if data == nil {
		return nil, errors.New("expense record data for update is nil")
	}

	data.UpdatedAt = time.Now()
	// Ensure ID is not changed and UserID is consistent if those are rules.
	// data.ID = id // Not needed if ID is part of the data struct and matches `id` argument.

	toMap, _ := utils.StructToMap(data)

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, err
	}

	err = r.DB.Update(ctx, id, toMap, *collection)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// DeleteExpenseRecord removes an expense record from the database.
func (r *ExpenseRecordRepository) DeleteExpenseRecord(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is empty for delete")
	}

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return err
	}

	return r.DB.Delete(ctx, id, *collection)
}

func (r *ExpenseRecordRepository) convertToEntity(data ...interface{}) ([]entity_finance.ExpenseRecord, error) {
	if data == nil {
		return nil, errors.New("data is nil")
	}

	if len(data) == 0 {
		return nil, errors.New("data is empty")
	}

	var result []entity_finance.ExpenseRecord
	for _, item := range data {
		responseEntity := entity_finance.ExpenseRecord{}
		if responseMap, ok := item.(map[string]interface{}); ok {
			responseEntity.ConvertISO8601ToTime("DueDate", responseMap["DueDate"].(string))
			responseEntity.ConvertISO8601ToTime("PaymentDate", responseMap["PaymentDate"].(string))
			responseEntity.ConvertISO8601ToTime("CreatedAt", responseMap["CreatedAt"].(string))
			responseEntity.ConvertISO8601ToTime("UpdatedAt", responseMap["UpdatedAt"].(string))

			if recurrenceCount, ok := responseMap["RecurrenceCount"]; ok {
				if count, ok := recurrenceCount.(int); ok {
					responseEntity.RecurrenceCount = count
				} else {
					responseEntity.RecurrenceCount = 0
				}
			} else {
				responseEntity.RecurrenceCount = 0
				if recurrenceNumber, ok := responseMap["RecurrenceNumber"]; ok {
					if count, ok := recurrenceNumber.(int); ok {
						responseEntity.RecurrenceNumber = count
					} else {
						responseEntity.RecurrenceNumber = 0
					}
				} else {
					responseEntity.RecurrenceNumber = 0 // Default to 0 if not present
				}

				responseEntity = entity_finance.ExpenseRecord{
					ID:             responseMap["ID"].(string),
					Category:       responseMap["Category"].(string),
					Subcategory:    responseMap["Subcategory"].(string),
					Amount:         responseMap["Amount"].(float64),
					BankPaidFrom:   responseMap["BankPaidFrom"].(string),
					CustomBankName: responseMap["CustomBankName"].(string),
					Description:    responseMap["Description"].(string),
					IsRecurring:    responseMap["IsRecurring"].(bool),
					UserID:         responseMap["UserID"].(string),
				}

				result = append(result, responseEntity)
			}
		}
	}

	return result, nil
}
