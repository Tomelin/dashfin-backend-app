package repository_finance

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	"github.com/Tomelin/dashfin-backend-app/internal/core/repository"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

// IncomeRecordRepository handles database operations for IncomeRecords.
type IncomeRecordRepository struct {
	DB         database.FirebaseDBInterface
	collection string
}

// InitializeIncomeRecordRepository creates a new IncomeRecordRepository.
func InitializeIncomeRecordRepository(db database.FirebaseDBInterface) (entity_finance.IncomeRecordRepositoryInterface, error) {
	if db == nil {
		return nil, errors.New("database is nil for IncomeRecordRepository")
	}

	return &IncomeRecordRepository{
		DB:         db,
		collection: "incomes", // As per plan
	}, nil
}

// CreateIncomeRecord adds a new income record to the database.
func (r *IncomeRecordRepository) CreateIncomeRecord(ctx context.Context, data *entity_finance.IncomeRecord) (*entity_finance.IncomeRecord, error) {
	if data == nil {
		return nil, errors.New("income record data is nil")
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
		log.Println("[RESPONSE] Error unmarshalling income record:", err)
		return nil, err
	}

	// If the response is a map, we can convert it back to IncomeRecord
	var responseEntity entity_finance.IncomeRecord
	if responseMap, ok := response.(map[string]interface{}); ok {

		log.Println("[RESPONSE] Income record created successfully:", responseMap)
		responseEntity.ConvertISO8601ToTime("ReceiptDate", responseMap["ReceiptDate"].(string))
		responseEntity.ConvertISO8601ToTime("CreatedAt", responseMap["CreatedAt"].(string))
		responseEntity.ConvertISO8601ToTime("UpdatedAt", responseMap["UpdatedAt"].(string))

		if recorrenceCount, ok := responseMap["RecurrenceCount"]; ok {
			if count, ok := recorrenceCount.(int); ok {
				responseEntity.RecurrenceCount = count
			} else {
				responseEntity.RecurrenceCount = 0
			}
		} else {
			responseEntity.RecurrenceCount = 0 // Default to 0 if not present
		}

		responseEntity = entity_finance.IncomeRecord{
			ID:            responseMap["ID"].(string),
			UserID:        responseMap["UserID"].(string),
			Description:   responseMap["Description"].(string),
			Category:      responseMap["Category"].(string),
			Amount:        responseMap["Amount"].(float64),
			BankAccountID: responseMap["BankAccountID"].(string),
			IsRecurring:   responseMap["IsRecurring"].(bool),
			Observations:  responseMap["Observations"].(string),
		}
	}

	return &responseEntity, nil
}

// GetIncomeRecordByID retrieves an income record by its ID.
// It also ensures that the record belongs to the UserID provided in the context if available.
func (r *IncomeRecordRepository) GetIncomeRecordByID(ctx context.Context, id string) (*entity_finance.IncomeRecord, error) {
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

	var records []entity_finance.IncomeRecord
	if err := json.Unmarshal(docs, &records); err != nil {
		return nil, err
	}

	var result *entity_finance.IncomeRecord
	for _, v := range records {
		if v.ID == filters["id"] {
			result = &v
			break
		}
	}

	if result == nil {
		return nil, errors.New("income record not found")
	}

	return result, nil
}

// // GetIncomeRecords retrieves income records based on query parameters.
func (r *IncomeRecordRepository) GetIncomeRecords(ctx context.Context, params *entity_finance.GetIncomeRecordsQueryParameters) ([]entity_finance.IncomeRecord, error) {
	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, err
	}

	result, err := r.DB.Get(ctx, *collection)
	if err != nil {
		return nil, err
	}

	var records []entity_finance.IncomeRecord
	if err := json.Unmarshal(result, &records); err != nil {
		return nil, err
	}

	return records, nil
}

// GetExpenseRecordsByFilter retrieves expense records based on a filter.
func (r *IncomeRecordRepository) GetExpenseRecordsByFilter(ctx context.Context, filter map[string]interface{}) ([]entity_finance.IncomeRecord, error) {
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

	var records []entity_finance.IncomeRecord
	if err := json.Unmarshal(result, &records); err != nil {
		return nil, err
	}
	return records, nil
}

// UpdateIncomeRecord updates an existing income record.
func (r *IncomeRecordRepository) UpdateIncomeRecord(ctx context.Context, id string, data *entity_finance.IncomeRecord) (*entity_finance.IncomeRecord, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("id is empty for update")
	}
	if data == nil {
		return nil, errors.New("income record data for update is nil")
	}

	data.UpdatedAt = time.Now()

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

// DeleteIncomeRecord removes an income record from the database.
// It must verify that the UserID in context matches the UserID of the record.
func (r *IncomeRecordRepository) DeleteIncomeRecord(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("id is empty for delete")
	}

	userIDFromCtx := ctx.Value("UserID")
	if userIDFromCtx == nil {
		return errors.New("userID not found in context for delete operation")
	}

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return err
	}

	return r.DB.Delete(ctx, id, *collection)
}
