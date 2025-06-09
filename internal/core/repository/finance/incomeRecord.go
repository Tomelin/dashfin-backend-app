package repository_finance

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	"github.com/Tomelin/dashfin-backend-app/internal/core/repository"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	// Ensure UserID is present, critical for data ownership
	if strings.TrimSpace(data.UserID) == "" {
		return nil, errors.New("userID is required in income record data")
	}

	// Validate the data before attempting to save
	if err := data.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed for income record: %w", err)
	}

	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()

	toMap, err := utils.StructToMap(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert income record to map: %w", err)
	}

	// Remove empty ID if present, Firestore generates it
	delete(toMap, "id")


	collectionName, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, fmt.Errorf("failed to set collection: %w", err)
	}

	// Use the underlying Firestore client from the DB interface
	client, ok := r.DB.GetClient().(*firestore.Client)
	if !ok {
		return nil, errors.New("failed to get Firestore client")
	}

	docRef, _, err := client.Collection(*collectionName).Add(ctx, toMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create income record in Firestore: %w", err)
	}
	data.ID = docRef.ID // Set the auto-generated ID

	// Return the data with ID, CreatedAt, UpdatedAt populated
	return data, nil
}

// GetIncomeRecordByID retrieves an income record by its ID.
// It also ensures that the record belongs to the UserID provided in the context if available.
func (r *IncomeRecordRepository) GetIncomeRecordByID(ctx context.Context, id string) (*entity_finance.IncomeRecord, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("id is empty")
	}

	collectionName, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, fmt.Errorf("failed to set collection: %w", err)
	}

	client, ok := r.DB.GetClient().(*firestore.Client)
	if !ok {
		return nil, errors.New("failed to get Firestore client")
	}

	docSnap, err := client.Collection(*collectionName).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errors.New("income record not found")
		}
		return nil, fmt.Errorf("failed to get income record from Firestore: %w", err)
	}

	var record entity_finance.IncomeRecord
	if err := docSnap.DataTo(&record); err != nil {
		return nil, fmt.Errorf("failed to map Firestore document to IncomeRecord: %w", err)
	}
	record.ID = docSnap.Ref.ID

	// Authorization check: If UserID is in context, verify ownership.
	// This is a repository-level check; service layer should also enforce this.
	userIDFromCtx := ctx.Value("UserID")
	if userIDFromCtx != nil {
		userIDStr, ok := userIDFromCtx.(string)
		if !ok {
			// Handle error: UserID in context is not a string
			return nil, errors.New("userID in context is not of type string")
		}
		if record.UserID != userIDStr {
			// This case should ideally be handled by query scopes, but as a safeguard:
			return nil, errors.New("income record not found (unauthorized access attempt)")
		}
	}


	return &record, nil
}

// GetIncomeRecords retrieves income records based on query parameters.
func (r *IncomeRecordRepository) GetIncomeRecords(ctx context.Context, params *entity_finance.GetIncomeRecordsQueryParameters) ([]entity_finance.IncomeRecord, error) {
	if params == nil {
		return nil, errors.New("query parameters are nil")
	}
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query parameters: %w", err)
	}
	// UserID must be present in params for filtering
	if strings.TrimSpace(params.UserID) == "" {
		return nil, errors.New("userID is required in query parameters")
	}

	collectionName, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, fmt.Errorf("failed to set collection: %w", err)
	}

	client, ok := r.DB.GetClient().(*firestore.Client)
	if !ok {
		return nil, errors.New("failed to get Firestore client")
	}

	query := client.Collection(*collectionName).Where("userId", "==", params.UserID)

	if params.Description != nil && *params.Description != "" {
		// Firestore doesn't support partial string matches (LIKE) directly in a simple query.
		// For a production system, you'd typically use a search service like Algolia, Elasticsearch,
		// or Firestore's own (limited) text search capabilities if enabled and suitable.
		// As a basic workaround, one might fetch and filter in memory, but this is not scalable.
		// For this implementation, we'll assume 'description' is for exact match or not implement if too complex for now.
		// Let's log a warning and proceed without description filtering if it's complex.
		// For now, we will filter by exact match if description is provided.
        query = query.Where("description", "==", *params.Description)
	}

	if params.StartDate != nil && *params.StartDate != "" {
		query = query.Where("receiptDate", ">=", *params.StartDate)
	}
	if params.EndDate != nil && *params.EndDate != "" {
		query = query.Where("receiptDate", "<=", *params.EndDate)
	}

	if params.SortKey != nil && *params.SortKey != "" {
		direction := firestore.Asc
		if params.SortDirection != nil && strings.ToLower(*params.SortDirection) == "desc" {
			direction = firestore.Desc
		}
		// Ensure the field name matches the Firestore document field name (bson or json tag if not transformed)
		// e.g. "receiptDate", "category", "amount"
		firestoreKey := *params.SortKey
		if firestoreKey == "receiptDate" { // common sort keys
			query = query.OrderBy("receiptDate", direction)
		} else if firestoreKey == "category" {
			query = query.OrderBy("category", direction)
		} else if firestoreKey == "amount" {
			query = query.OrderBy("amount", direction)
		} else {
            // If an unknown sortKey is provided, we might default to a sort or return an error.
            // For now, let's default to sorting by receiptDate if sortKey is invalid or not specified.
            query = query.OrderBy("receiptDate", firestore.Desc) // Default sort
        }
	} else {
		// Default sort if no key is provided
		query = query.OrderBy("receiptDate", firestore.Desc)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	var records []entity_finance.IncomeRecord
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate income records: %w", err)
		}

		var record entity_finance.IncomeRecord
		if err := doc.DataTo(&record); err != nil {
			// Log error and continue if one record fails, or return error immediately
			return nil, fmt.Errorf("failed to map Firestore document to IncomeRecord: %w", err)
		}
		record.ID = doc.Ref.ID
		records = append(records, record)
	}

	if len(records) == 0 {
		// Return empty slice, not an error, if no records found matching criteria
		return []entity_finance.IncomeRecord{}, nil
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
	if strings.TrimSpace(data.UserID) == "" {
		return nil, errors.New("userID is required in income record data for update")
	}

	// Validate the incoming data
	if err := data.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed for income record update: %w", err)
	}

	collectionName, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return nil, fmt.Errorf("failed to set collection: %w", err)
	}

	client, ok := r.DB.GetClient().(*firestore.Client)
	if !ok {
		return nil, errors.New("failed to get Firestore client")
	}

	docRef := client.Collection(*collectionName).Doc(id)

	// Check if the document exists and belongs to the user before updating
	// This is an important step for authorization
	existingDoc, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errors.New("income record not found for update")
		}
		return nil, fmt.Errorf("failed to get existing income record: %w", err)
	}
	var existingRecord entity_finance.IncomeRecord
	if err := existingDoc.DataTo(&existingRecord); err != nil {
		return nil, fmt.Errorf("failed to map existing Firestore document: %w", err)
	}
	if existingRecord.UserID != data.UserID {
		return nil, errors.New("authorization failed: user does not own this income record")
	}


	data.UpdatedAt = time.Now()
	data.ID = id // Ensure ID is the one from path param
	// CreatedAt should not change on update, so we can preserve it from existing record if needed
	data.CreatedAt = existingRecord.CreatedAt


	toMap, err := utils.StructToMap(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert income record to map for update: %w", err)
	}

	// Remove ID from map as it's part of the Doc path, and to prevent trying to update _id
	delete(toMap, "id")

	if _, err := docRef.Set(ctx, toMap, firestore.MergeAll); err != nil { // Use MergeAll to only update provided fields
		return nil, fmt.Errorf("failed to update income record in Firestore: %w", err)
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

	userIDStr, ok := userIDFromCtx.(string)
	if !ok {
		return errors.New("userID in context is not of type string for delete operation")
	}
	if userIDStr == "" {
		return errors.New("userID in context is empty for delete operation")
	}


	collectionName, err := repository.SetCollection(ctx, r.collection)
	if err != nil {
		return fmt.Errorf("failed to set collection: %w", err)
	}

	client, ok := r.DB.GetClient().(*firestore.Client)
	if !ok {
		return errors.New("failed to get Firestore client")
	}

	docRef := client.Collection(*collectionName).Doc(id)

	// Before deleting, verify ownership
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return errors.New("income record not found for delete") // Or return nil if "not found" means "already deleted"
		}
		return fmt.Errorf("failed to get income record for delete verification: %w", err)
	}

	var record entity_finance.IncomeRecord
	if err := docSnap.DataTo(&record); err != nil {
		return fmt.Errorf("failed to map Firestore document for delete verification: %w", err)
	}

	if record.UserID != userIDStr {
		return errors.New("authorization failed: user does not own this income record for delete")
	}

	if _, err := docRef.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete income record from Firestore: %w", err)
	}

	return nil
}
