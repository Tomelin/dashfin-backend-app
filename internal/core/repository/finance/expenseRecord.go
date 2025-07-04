package repository_finance

import (
	"context"
	"encoding/json"
	"errors"
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

	repo := make([]interface{}, 1)
	repo[0] = response

	responseEntity, err := r.convertToEntity(repo)
	if err != nil {
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

	var response []interface{}
	err = json.Unmarshal(docs, &response)
	if err != nil {
		return nil, err
	}

	responseEntity, err := r.convertToEntity(response)
	if err != nil {
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

	var response []interface{}
	err = json.Unmarshal(result, &response)
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

	var response []interface{}
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, err
	}

	responseEntity, err := r.convertToEntity(response)
	if err != nil {
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

func (r *ExpenseRecordRepository) convertToEntity(data []interface{}) ([]entity_finance.ExpenseRecord, error) {

	if data == nil {
		return nil, errors.New("data is nil")
	}

	var result []entity_finance.ExpenseRecord
	for _, item := range data {

		if itemMap, ok := item.(map[string]interface{}); ok {
			record := entity_finance.ExpenseRecord{}
			if id, ok := itemMap["id"]; ok {
				record.ID = id.(string)
			} else {
				record.ID = itemMap["ID"].(string)
			}

			if category, ok := itemMap["Category"]; ok {
				record.Category = category.(string)
			} else {
				record.Category = itemMap["category"].(string)
			}

			if description, ok := itemMap["Description"]; ok {
				record.Description = description.(string)
			} else {
				if description, ok := itemMap["description"]; ok {
					record.Description = description.(string)
				} else {
					record.Description = ""
				}
			}

			if subcategory, ok := itemMap["Subcategory"]; ok {
				record.Subcategory = subcategory.(string)
			} else {
				if subcategory, ok := itemMap["subcategory"]; ok {
					record.Subcategory = subcategory.(string)
				} else {
					record.Subcategory = ""
				}
			}

			if amount, ok := itemMap["Amount"]; ok {
				record.Amount = amount.(float64)
			} else {
				if amount, ok := itemMap["amount"]; ok {
					record.Amount = amount.(float64)
				}
			}

			if userID, ok := itemMap["UserID"]; ok {
				record.UserID = userID.(string)
			} else {
				if userID, ok := itemMap["userId"]; ok {
					record.UserID = userID.(string)
				}
			}

			if customBankName, ok := itemMap["CustomBankName"]; ok {
				record.CustomBankName = customBankName.(string)
			} else {
				if customBankName, ok := itemMap["customBankName"]; ok {
					record.CustomBankName = customBankName.(string)
				}
			}

			if bankPaidFrom, ok := itemMap["BankPaidFrom"]; ok {
				record.BankPaidFrom = bankPaidFrom.(string)
			} else {
				if bankPaidFrom, ok := itemMap["bankPaidFrom"]; ok {
					record.BankPaidFrom = bankPaidFrom.(string)
				}
			}

			if isRecurring, ok := itemMap["IsRecurring"]; ok {
				record.IsRecurring = isRecurring.(bool)
			} else {
				if isRecurring, ok := itemMap["isRecurring"]; ok {
					record.IsRecurring = isRecurring.(bool)
				}
			}

			if dueDate, ok := itemMap["DueDate"]; ok {
				t, _ := time.Parse("2006-01-02T15:04:05Z", dueDate.(string))
				record.DueDate = t
			} else {
				if dueDate, ok := itemMap["dueDate"]; ok {
					t, _ := time.Parse("2006-01-02T15:04:05Z", dueDate.(string))
					record.DueDate = t
				}
			}

			if paymentDate, ok := itemMap["PaymentDate"]; ok {
				t, _ := time.Parse("2006-01-02T15:04:05Z", paymentDate.(string))
				record.PaymentDate = t
			} else {
				if paymentDate, ok := itemMap["paymentDate"]; ok {
					t, _ := time.Parse("2006-01-02T15:04:05Z", paymentDate.(string))
					record.PaymentDate = t
				}
			}

			if recurrenceCount, ok := itemMap["RecurrenceCount"]; ok {
				if count, ok := recurrenceCount.(float64); ok {
					record.RecurrenceCount = int(count)
				} else {
					record.RecurrenceCount = 0
				}
			} else {
				if recurrenceCount, ok := itemMap["recurrenceCount"]; ok {
					if count, ok := recurrenceCount.(float64); ok {
						record.RecurrenceCount = int(count)
					} else {
						record.RecurrenceCount = 0
					}
				}
			}

			if recurrenceNumber, ok := itemMap["RecurrenceNumber"]; ok {
				if count, ok := recurrenceNumber.(float64); ok {
					record.RecurrenceNumber = int(count)
				} else {
					record.RecurrenceNumber = 0
				}
			} else {
				if recurrenceNumber, ok := itemMap["recurrenceNumber"]; ok {
					if count, ok := recurrenceNumber.(float64); ok {
						record.RecurrenceNumber = int(count)
					} else {
						record.RecurrenceNumber = 0
					}
				}
			}

			result = append(result, record)
		}
	}
	return result, nil
}
