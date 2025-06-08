package service_finance

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
)

// ExpenseRecordService provides business logic for expense records.
type ExpenseRecordService struct {
	Repo entity_finance.ExpenseRecordRepositoryInterface
}

// InitializeExpenseRecordService creates a new ExpenseRecordService.
func InitializeExpenseRecordService(repo entity_finance.ExpenseRecordRepositoryInterface) (entity_finance.ExpenseRecordServiceInterface, error) {
	if repo == nil {
		return nil, errors.New("repository is nil for ExpenseRecordService")
	}
	return &ExpenseRecordService{Repo: repo}, nil
}

// CreateExpenseRecord handles the creation of a new expense record.
func (s *ExpenseRecordService) CreateExpenseRecord(ctx context.Context, data *entity_finance.ExpenseRecord) (*entity_finance.ExpenseRecord, error) {
	if data == nil {
		return nil, errors.New("expense record data is nil")
	}

	if err := data.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// UserID should be set by the handler from context before calling this service method.
	if data.UserID == "" {
		return nil, errors.New("userID is required in expense record data")
	}

	// For new records, ensure CreatedAt and UpdatedAt are set.
	// ID will be set by the repository.
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()

	// If the expense is recurring and RecurrenceCount is greater than zero,
	// we might need to handle the creation of future occurrences here or in the repository.
	if data.IsRecurring && data.RecurrenceCount > 0 {
		expensesCreated := make([]entity_finance.ExpenseRecord, 0)

		snapDueDate := data.DueDate
		for i := 1; i < data.RecurrenceCount; i++ {
			log.Println("i starts at ", i)
			if i == 1 {
				result, _ := s.Repo.CreateExpenseRecord(ctx, data)
				expensesCreated = append(expensesCreated, *result)
				s.Repo.CreateExpenseRecord(ctx, data)
			} else {
				parsedDueDate, _ := time.Parse("2006-01-02", snapDueDate)
				newDate := parsedDueDate.AddDate(0, i, 0) // Add i months
				data.DueDate = newDate.Format("2006-01-02")
				log.Println("for else and next date", data.DueDate)
				result, err := s.Repo.CreateExpenseRecord(ctx, data)
				log.Println("for if ", err)
				expensesCreated = append(expensesCreated, *result)
			}
		}
		log.Println("total...", len(expensesCreated))
		log.Println(expensesCreated)
		return &expensesCreated[0], nil // Return the first created expense record

	}

	return s.Repo.CreateExpenseRecord(ctx, data)
}

// GetExpenseRecordByID retrieves an expense record by its ID, ensuring user authorization.
func (s *ExpenseRecordService) GetExpenseRecordByID(ctx context.Context, id string) (*entity_finance.ExpenseRecord, error) {
	if id == "" {
		return nil, errors.New("id is empty")
	}

	record, err := s.Repo.GetExpenseRecordByID(ctx, id)
	if err != nil {
		return nil, err // Handles "not found" from repository
	}

	// Authorization: Ensure the record belongs to the user in context.
	// This assumes UserID is correctly populated in the record and available in context.
	userIDFromCtx := ctx.Value("UserID")
	if userIDFromCtx == nil || record.UserID != userIDFromCtx.(string) {
		// Log this attempt, could be a security issue or bug
		// log.Printf("Authorization failed: User %v attempted to access record %s owned by %s", userIDFromCtx, id, record.UserID)
		return nil, errors.New("expense record not found or access denied") // Generic message for security
	}

	return record, nil
}

// GetExpenseRecords retrieves all expense records for the authenticated user.
func (s *ExpenseRecordService) GetExpenseRecords(ctx context.Context) ([]entity_finance.ExpenseRecord, error) {
	userIDFromCtx := ctx.Value("UserID")
	if userIDFromCtx == nil || userIDFromCtx.(string) == "" {
		return nil, errors.New("userID not found in context")
	}

	filter := map[string]interface{}{
		"userId": userIDFromCtx.(string),
	}

	records, err := s.Repo.GetExpenseRecordsByFilter(ctx, filter)
	if err != nil {
		// If the error is because no records were found, it might be better to return an empty slice and no error.
		// This depends on the desired API contract. For now, mirroring BankAccount which returns "not found".
		return nil, err
	}
	// if len(records) == 0 { // This check might be redundant if repo returns specific "not found" error
	// 	return nil, errors.New("no expense records found for the user")
	// }
	return records, nil
}

func (s *ExpenseRecordService) GetExpenseRecordsByDate(ctx context.Context, filter *entity_finance.ExpenseRecordQueryByDate) ([]entity_finance.ExpenseRecord, error) {
	userIDFromCtx := ctx.Value("UserID")
	if userIDFromCtx == nil || userIDFromCtx.(string) == "" {
		return nil, errors.New("userID not found in context")
	}

	records, err := s.Repo.GetExpenseRecords(ctx)
	if err != nil {
		return nil, err
	}

	var filteredRecords []entity_finance.ExpenseRecord
	startDate, _ := time.Parse("2006-01-02", filter.StartDate)
	endDate, _ := time.Parse("2006-01-02", filter.EndDate)

	for _, record := range records {
		parsedDueDate, err := time.Parse("2006-01-02", record.DueDate)
		if err != nil {
			return nil, errors.New("dueDate must be in ISO 8601 format (YYYY-MM-DD)")
		}
		if startDate != (time.Time{}) && parsedDueDate.Before(startDate) {
			continue
		}

		if endDate != (time.Time{}) && record.CreatedAt.After(endDate) {
			continue
		}
		filteredRecords = append(filteredRecords, record)
	}

	return filteredRecords, nil
}

// GetExpenseRecordsByFilter retrieves expense records based on a filter for the authenticated user.
func (s *ExpenseRecordService) GetExpenseRecordsByFilter(ctx context.Context, filter map[string]interface{}) ([]entity_finance.ExpenseRecord, error) {
	if filter == nil {
		return nil, errors.New("filter data is nil")
	}
	userIDFromCtx := ctx.Value("UserID")
	if userIDFromCtx == nil || userIDFromCtx.(string) == "" {
		return nil, errors.New("userID not found in context")
	}

	// Ensure the filter always includes the UserID from context to scope results.
	filter["userId"] = userIDFromCtx.(string)

	records, err := s.Repo.GetExpenseRecordsByFilter(ctx, filter)
	if err != nil {
		return nil, err
	}
	// if len(records) == 0 { // Similar to GetExpenseRecords, decide if "not found" is an error or empty slice.
	// 	return nil, errors.New("no expense records found matching the filter for the user")
	// }
	return records, nil
}

// UpdateExpenseRecord handles updating an existing expense record.
func (s *ExpenseRecordService) UpdateExpenseRecord(ctx context.Context, id string, data *entity_finance.ExpenseRecord) (*entity_finance.ExpenseRecord, error) {
	if id == "" {
		return nil, errors.New("id is empty for update")
	}
	if data == nil {
		return nil, errors.New("expense record data for update is nil")
	}

	if err := data.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed for update: %w", err)
	}

	// UserID should be set by the handler from context before calling this service method.
	// It should match the original record's UserID and the UserID in context.
	userIDFromCtx := ctx.Value("UserID")
	if userIDFromCtx == nil || data.UserID != userIDFromCtx.(string) {
		return nil, errors.New("user ID mismatch or not found in context for update")
	}

	// First, retrieve the existing record to ensure it exists and belongs to the user.
	existingRecord, err := s.Repo.GetExpenseRecordByID(ctx, id)
	if err != nil {
		return nil, err // Handles "not found" from repository
	}

	if existingRecord.UserID != data.UserID { // Also check against UserID from context
		// log.Printf("Authorization failed for update: User %v attempted to update record %s owned by %s", data.UserID, id, existingRecord.UserID)
		return nil, errors.New("expense record not found or access denied for update")
	}

	// Ensure critical fields like ID and UserID are not changed by the update payload directly,
	// or are consistent.
	data.ID = existingRecord.ID               // Preserve original ID
	data.UserID = existingRecord.UserID       // Preserve original UserID
	data.CreatedAt = existingRecord.CreatedAt // Preserve original CreatedAt
	data.UpdatedAt = time.Now()               // Update timestamp

	return s.Repo.UpdateExpenseRecord(ctx, id, data)
}

// DeleteExpenseRecord handles deleting an expense record.
func (s *ExpenseRecordService) DeleteExpenseRecord(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is empty for delete")
	}

	userIDFromCtx := ctx.Value("UserID")
	if userIDFromCtx == nil || userIDFromCtx.(string) == "" {
		return errors.New("userID not found in context for delete")
	}

	// Retrieve the record first to ensure it belongs to the user.
	recordToVerify, err := s.Repo.GetExpenseRecordByID(ctx, id)
	if err != nil {
		return err // Handles "not found" from repository, which is fine.
	}

	if recordToVerify.UserID != userIDFromCtx.(string) {
		// log.Printf("Authorization failed for delete: User %v attempted to delete record %s owned by %s", userIDFromCtx, id, recordToVerify.UserID)
		return errors.New("expense record not found or access denied for delete")
	}

	return s.Repo.DeleteExpenseRecord(ctx, id)
}
