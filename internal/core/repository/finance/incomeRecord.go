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

	repo := make([]interface{}, 1)
	repo[0] = response
	// If the response is a map, we can convert it back to IncomeRecord

	responseEntity, err := r.convertToEntity(repo)
	if err != nil {
		return nil, err
	}

	return &responseEntity[0], nil
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

	var response []interface{}
	err = json.Unmarshal(docs, &response)
	if err != nil {
		return nil, err
	}

	responseEntity, err := r.convertToEntity(response)
	if err != nil {
		return nil, err
	}

	var result *entity_finance.IncomeRecord
	for _, v := range responseEntity {
		if v.ID == filters["id"] {
			result = &v
			break
		}
	}
	return result, nil
}

// // GetIncomeRecords retrieves income records based on query parameters.
func (r *IncomeRecordRepository) GetIncomeRecords(ctx context.Context, params *entity_finance.GetIncomeRecordsQueryParameters) ([]entity_finance.IncomeRecord, error) {
	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, err
	}

	docs, err := r.DB.Get(ctx, *collection)
	if err != nil {
		return nil, err
	}

	var response []interface{}
	err = json.Unmarshal(docs, &response)
	if err != nil {
		return nil, err
	}

	responseEntity, err := r.convertToEntity(response)
	if err != nil {
		return nil, err
	}

	return responseEntity, nil
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

	docs, err := r.DB.GetByFilter(ctx, filter, *collection)
	if err != nil {
		return nil, err
	}

	var response []interface{}
	err = json.Unmarshal(docs, &response)
	if err != nil {
		return nil, err
	}

	responseEntity, err := r.convertToEntity(response)
	if err != nil {
		return nil, err
	}
	return responseEntity, nil
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

func (r *IncomeRecordRepository) convertToEntity(data []interface{}) ([]entity_finance.IncomeRecord, error) {

	if data == nil {
		return nil, errors.New("data is nil")
	}

	var result []entity_finance.IncomeRecord
	for _, item := range data {
		if itemMap, ok := item.(map[string]interface{}); ok {
			record := entity_finance.IncomeRecord{}

			if userID, ok := itemMap["UserID"]; ok {
				record.UserID = userID.(string)
			} else {
				record.UserID = itemMap["userId"].(string)
			}

			if id, ok := itemMap["id"]; ok {
				record.ID = id.(string)
			} else {
				record.ID = itemMap["ID"].(string)
			}
			if description, ok := itemMap["Description"]; ok {
				record.Description = description.(string)
			} else {
				record.Description = itemMap["description"].(string)
			}

			if category, ok := itemMap["Category"]; ok {
				record.Category = category.(string)
			} else {
				record.Category = itemMap["category"].(string)
			}
			if amount, ok := itemMap["Amount"]; ok {
				record.Amount = amount.(float64)
			} else {
				record.Amount = itemMap["amount"].(float64)
			}
			if bankAccountID, ok := itemMap["BankAccountID"]; ok {
				record.BankAccountID = bankAccountID.(string)
			} else {
				record.BankAccountID = itemMap["bankAccountId"].(string)
			}
			if isRecurring, ok := itemMap["IsRecurring"]; ok {
				record.IsRecurring = isRecurring.(bool)
			} else {
				record.IsRecurring = itemMap["isRecurring"].(bool)
			}
			if observations, ok := itemMap["Observations"]; ok {
				record.Observations = observations.(string)
			} else {
				record.Observations = itemMap["observations"].(string)
			}

			// record.ConvertISO8601ToTime("CreatedAt", itemMap["CreatedAt"].(string))
			// record.ConvertISO8601ToTime("UpdatedAt", itemMap["UpdatedAt"].(string))

			var receiptDate string
			if receiptDateValue, ok := itemMap["ReceiptDate"]; ok {
				receiptDate = receiptDateValue.(string)
			} else {
				receiptDate = itemMap["receiptDate"].(string)
			}

			t, _ := time.Parse("2006-01-02T15:04:05Z", receiptDate)
			record.ReceiptDate = t

			if recurrenceCount, ok := itemMap["RecurrenceCount"]; ok {
				if count, ok := recurrenceCount.(int); ok {
					record.RecurrenceCount = count
				} else {
					record.RecurrenceCount = 0 // Default to 0 if not present
				}
			} else {
				record.RecurrenceCount = 0 // Default to 0 if not present
			}

			if recurrenceNumber, ok := itemMap["RecurrenceNumber"]; ok {
				if number, ok := recurrenceNumber.(int); ok {
					record.RecurrenceNumber = number
				} else {
					record.RecurrenceNumber = 0 // Default to 0 if not present
				}
			} else {
				record.RecurrenceNumber = 0 // Default to 0 if not present
			}
			result = append(result, record)
		}
	}

	return result, nil
}
