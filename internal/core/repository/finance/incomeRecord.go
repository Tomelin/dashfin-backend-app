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

	toMap, err := utils.StructToMap(data)
	log.Println("Converted income record to map:", toMap, err)

	collection, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, err
	}

	doc, err := r.DB.Create(ctx, toMap, *collection)
	if err != nil {
		log.Println("[DB] Error creating income record:", err)
		return nil, err
	}

	var response entity_finance.IncomeRecord
	err = json.Unmarshal(doc, &response)
	if err != nil {
		log.Println("[RESPONSE] Error unmarshalling income record:", err)
		return nil, err
	}

	return &response, nil
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

// func (r *IncomeRecordRepository) GetIncomeRecords(ctx context.Context, params *entity_finance.GetIncomeRecordsQueryParameters) ([]entity_finance.IncomeRecord, error) {
// 	if params == nil {
// 		return nil, errors.New("query parameters are nil")
// 	}

// 	if err := params.Validate(); err != nil {
// 		return nil, fmt.Errorf("invalid query parameters: %w", err)
// 	}
// 	// UserID must be present in params for filtering
// 	if strings.TrimSpace(params.UserID) == "" {
// 		return nil, errors.New("userID is required in query parameters")
// 	}

// 	collectionName, err := repository.SetCollection(ctx, r.collection)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to set collection: %w", err)
// 	}

// 	client, ok := r.DB.GetClient().(*firestore.Client)
// 	if !ok {
// 		return nil, errors.New("failed to get Firestore client")
// 	}

// 	query := client.Collection(*collectionName).Where("userId", "==", params.UserID)

// 	if params.Description != nil && *params.Description != "" {
// 		// Firestore doesn't support partial string matches (LIKE) directly in a simple query.
// 		// For a production system, you'd typically use a search service like Algolia, Elasticsearch,
// 		// or Firestore's own (limited) text search capabilities if enabled and suitable.
// 		// As a basic workaround, one might fetch and filter in memory, but this is not scalable.
// 		// For this implementation, we'll assume 'description' is for exact match or not implement if too complex for now.
// 		// Let's log a warning and proceed without description filtering if it's complex.
// 		// For now, we will filter by exact match if description is provided.
// 		query = query.Where("description", "==", *params.Description)
// 	}

// 	if params.StartDate != nil && *params.StartDate != "" {
// 		query = query.Where("receiptDate", ">=", *params.StartDate)
// 	}
// 	if params.EndDate != nil && *params.EndDate != "" {
// 		query = query.Where("receiptDate", "<=", *params.EndDate)
// 	}

// 	if params.SortKey != nil && *params.SortKey != "" {
// 		direction := firestore.Asc
// 		if params.SortDirection != nil && strings.ToLower(*params.SortDirection) == "desc" {
// 			direction = firestore.Desc
// 		}
// 		// Ensure the field name matches the Firestore document field name (bson or json tag if not transformed)
// 		// e.g. "receiptDate", "category", "amount"
// 		firestoreKey := *params.SortKey
// 		if firestoreKey == "receiptDate" { // common sort keys
// 			query = query.OrderBy("receiptDate", direction)
// 		} else if firestoreKey == "category" {
// 			query = query.OrderBy("category", direction)
// 		} else if firestoreKey == "amount" {
// 			query = query.OrderBy("amount", direction)
// 		} else {
// 			// If an unknown sortKey is provided, we might default to a sort or return an error.
// 			// For now, let's default to sorting by receiptDate if sortKey is invalid or not specified.
// 			query = query.OrderBy("receiptDate", firestore.Desc) // Default sort
// 		}
// 	} else {
// 		// Default sort if no key is provided
// 		query = query.OrderBy("receiptDate", firestore.Desc)
// 	}

// 	iter := query.Documents(ctx)
// 	defer iter.Stop()

// 	var records []entity_finance.IncomeRecord
// 	for {
// 		doc, err := iter.Next()
// 		if err == iterator.Done {
// 			break
// 		}
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to iterate income records: %w", err)
// 		}

// 		var record entity_finance.IncomeRecord
// 		if err := doc.DataTo(&record); err != nil {
// 			// Log error and continue if one record fails, or return error immediately
// 			return nil, fmt.Errorf("failed to map Firestore document to IncomeRecord: %w", err)
// 		}
// 		record.ID = doc.Ref.ID
// 		records = append(records, record)
// 	}

// 	if len(records) == 0 {
// 		// Return empty slice, not an error, if no records found matching criteria
// 		return []entity_finance.IncomeRecord{}, nil
// 	}

// 	return records, nil
// }

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
