package repository_finance

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
	"github.com/google/uuid"
)

// ExpenseRecordRepository handles database operations for ExpenseRecords.
type ExpenseRecordRepository struct {
	DB database.FirebaseDBInterface
}

// InitializeExpenseRecordRepository creates a new ExpenseRecordRepository.
func InitializeExpenseRecordRepository(db database.FirebaseDBInterface) (entity_finance.ExpenseRecordRepositoryInterface, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}
	return &ExpenseRecordRepository{DB: db}, nil
}

// CreateExpenseRecord adds a new expense record to the database.
func (r *ExpenseRecordRepository) CreateExpenseRecord(ctx context.Context, data *entity_finance.ExpenseRecord) (*entity_finance.ExpenseRecord, error) {
	if data == nil {
		return nil, errors.New("expense record data is nil")
	}

	// Generate ID and set timestamps
	data.ID = uuid.NewString()
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()

	_, err := r.DB.Create(ctx, data, "expenseRecords")
	if err != nil {
		return nil, err
	}
	return data, nil
}

// GetExpenseRecordByID retrieves an expense record by its ID.
func (r *ExpenseRecordRepository) GetExpenseRecordByID(ctx context.Context, id string) (*entity_finance.ExpenseRecord, error) {
	if id == "" {
		return nil, errors.New("id is empty")
	}

	filters := map[string]interface{}{
		"_id": id, // Assuming MongoDB/Firebase uses _id or similar for document ID
	}

	result, err := r.DB.GetByFilter(ctx, filters, "expenseRecords")
	if err != nil {
		return nil, err
	}

	var records []entity_finance.ExpenseRecord
	if err := json.Unmarshal(result, &records); err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, errors.New("expense record not found")
	}
	return &records[0], nil
}

// GetExpenseRecords retrieves all expense records for the user in context (or all if no user context).
func (r *ExpenseRecordRepository) GetExpenseRecords(ctx context.Context) ([]entity_finance.ExpenseRecord, error) {
	// Potentially add filtering by UserID from context if available
	// For now, gets all records.
	// userID := ctx.Value("UserID").(string)
	// filters := map[string]interface{}{
	//  "userId": userID,
	// }
	// result, err := r.DB.GetByFilter(ctx, filters, "expenseRecords")

	result, err := r.DB.Get(ctx, "expenseRecords")
	if err != nil {
		return nil, err
	}

	var records []entity_finance.ExpenseRecord
	if err := json.Unmarshal(result, &records); err != nil {
		return nil, err
	}
	return records, nil
}

// GetExpenseRecordsByFilter retrieves expense records based on a filter.
func (r *ExpenseRecordRepository) GetExpenseRecordsByFilter(ctx context.Context, filter map[string]interface{}) ([]entity_finance.ExpenseRecord, error) {
	if filter == nil {
		return nil, errors.New("filter data is nil")
	}

	// Ensure UserID from context is part of the filter if not admin
	// userID := ctx.Value("UserID").(string)
	// filter["userId"] = userID


	result, err := r.DB.GetByFilter(ctx, filter, "expenseRecords")
	if err != nil {
		return nil, err
	}

	var records []entity_finance.ExpenseRecord
	if err := json.Unmarshal(result, &records); err != nil {
		return nil, err
	}
	return records, nil
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

	err := r.DB.Update(ctx, id, data, "expenseRecords")
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
	return r.DB.Delete(ctx, id, "expenseRecords")
}
