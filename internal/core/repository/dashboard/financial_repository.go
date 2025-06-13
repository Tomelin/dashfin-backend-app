package dashboard

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
	"google.golang.org/api/iterator"
	// "github.com/Tomelin/dashfin-backend-app/pkg/database" // No longer needed directly here
)

// FinancialRepositoryInterface defines the contract for accessing financial data
// from a persistent store, specifically for dashboard-related information such as
// expense plans, actual expenses, and category definitions.
type FinancialRepositoryInterface interface {
	// GetExpensePlanning retrieves a specific expense planning document for a given user, month, and year.
	// It returns the ExpensePlanningDoc if found. If no document exists for the criteria,
	// it returns (nil, nil) to indicate not found without an application error.
	// Errors during database interaction are returned as a non-nil error.
	GetExpensePlanning(ctx context.Context, client *firestore.Client, userID string, month, year int) (*entity_dashboard.ExpensePlanningDoc, error)

	// GetExpenses retrieves all 'paid' expense documents for a given user that fall within
	// the specified month and year.
	// It returns a slice of ExpenseDoc. If no expenses match, an empty slice is returned.
	// Errors during database interaction are returned as a non-nil error.
	GetExpenses(ctx context.Context, client *firestore.Client, userID string, month, year int) ([]entity_dashboard.ExpenseDoc, error)

	// GetExpenseCategories retrieves all active expense category definitions.
	// These definitions provide metadata like labels for category keys.
	// It returns a slice of ExpenseCategoryDoc. If no categories are defined or active,
	// an empty slice is returned.
	// Errors during database interaction are returned as a non-nil error.
	GetExpenseCategories(ctx context.Context, client *firestore.Client) ([]entity_dashboard.ExpenseCategoryDoc, error)
}

// FirebaseFinancialRepository implements the FinancialRepositoryInterface using Google Firestore
// as the data source.
type FirebaseFinancialRepository struct {
	// This struct is intentionally empty as the Firestore client is passed directly to its methods.
	// This approach can be useful for managing client lifecycles or transactions at a higher level (e.g., in the service layer).
}

// NewFirebaseFinancialRepository creates and returns a new instance of FirebaseFinancialRepository.
// It implements the FinancialRepositoryInterface.
func NewFirebaseFinancialRepository() FinancialRepositoryInterface {
	return &FirebaseFinancialRepository{}
}

// GetExpensePlanning retrieves a specific expense planning document for a user, month, and year from Firestore.
// The document ID is constructed as "userID_year_month".
// If the document is not found, it returns (nil, nil) indicating no data, not an error.
func (r *FirebaseFinancialRepository) GetExpensePlanning(ctx context.Context, client *firestore.Client, userID string, month, year int) (*entity_dashboard.ExpensePlanningDoc, error) {
	docID := fmt.Sprintf("%s_%d_%d", userID, year, month)
	docSnap, err := client.Collection("expenses_planning").Doc(docID).Get(ctx)
	if err != nil {
		if err.Error() == "rpc error: code = NotFound desc = document not found" || iterator.Done.Error() == err.Error() || err.Error() == "datastore: no such entity" { // Common ways NotFound is represented
			// Check if the error indicates "not found". Firestore might return an error with code NotFound.
			// For GRPC, the error might be something like "rpc error: code = NotFound desc = ..."
			// It's safer to check specific error codes if available, but string matching is a common fallback.
			// For now, a simple string check for "not found" or iterator.Done will be used.
			// A more robust solution would involve checking `status.Code(err) == codes.NotFound`.
			return nil, nil // Document not found is not treated as an application error here.
		}
		return nil, fmt.Errorf("failed to get expense planning document %s: %w", docID, err)
	}

	var planningDoc entity_dashboard.ExpensePlanningDoc
	if err := docSnap.DataTo(&planningDoc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal expense planning document %s into ExpensePlanningDoc: %w", docID, err)
	}
	planningDoc.ID = docSnap.Ref.ID // Populate ID from the document reference.
	return &planningDoc, nil
}

// GetExpenses retrieves all 'paid' expenses for a given user within a specific month and year from Firestore.
// It filters expenses by 'user_id', 'status' ("paid"), and a 'payment_date' range corresponding
// to the start and end of the provided month and year.
func (r *FirebaseFinancialRepository) GetExpenses(ctx context.Context, client *firestore.Client, userID string, month, year int) ([]entity_dashboard.ExpenseDoc, error) {
	var expenses []entity_dashboard.ExpenseDoc

	// Defines the period for filtering expenses.
	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0) // End of the month is the start of the next month (exclusive).

	query := client.Collection("expenses").
		Where("user_id", "==", userID).      // Filter by user ID.
		Where("status", "==", "paid").         // Filter by status "paid".
		Where("payment_date", ">=", startOfMonth). // Filter by payment date range.
		Where("payment_date", "<", endOfMonth)

	iter := query.Documents(ctx)
	defer iter.Stop()

	for {
		docSnap, err := iter.Next()
		if err == iterator.Done { // End of iteration.
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over expense documents: %w", err)
		}

		var expense entity_dashboard.ExpenseDoc
		if err := docSnap.DataTo(&expense); err != nil {
			// Log or handle individual document unmarshalling errors if necessary.
			// Skipping corrupted documents and continuing might be an option depending on requirements.
			fmt.Printf("Warning: failed to unmarshal expense document %s into ExpenseDoc: %v\n", docSnap.Ref.ID, err)
			continue
		}
		expense.ID = docSnap.Ref.ID // Populate ID.
		expenses = append(expenses, expense)
	}

	return expenses, nil
}

// GetExpenseCategories retrieves all active expense categories from the "expense_categories" collection in Firestore.
// It filters categories by the 'active' field being true.
// Returns an empty slice if the collection is empty, no categories are active, or in case of non-critical errors like collection not found.
// Actual database interaction errors will be returned.
func (r *FirebaseFinancialRepository) GetExpenseCategories(ctx context.Context, client *firestore.Client) ([]entity_dashboard.ExpenseCategoryDoc, error) {
	var categories []entity_dashboard.ExpenseCategoryDoc

	query := client.Collection("expense_categories").Where("active", "==", true) // Filter by active status.
	iter := query.Documents(ctx)
	defer iter.Stop()

	for {
		docSnap, err := iter.Next()
		if err == iterator.Done { // End of iteration.
			break
		}
		if err != nil {
			// Firestore might not error on Documents() for a non-existent collection but could error on Next().
			// Depending on strictness, one might check for specific "not found" codes here too,
			// but generally, iteration errors are treated as issues.
			return nil, fmt.Errorf("failed to iterate over expense category documents: %w", err)
		}

		var category entity_dashboard.ExpenseCategoryDoc
		if err := docSnap.DataTo(&category); err != nil {
			fmt.Printf("Warning: failed to unmarshal expense category document %s into ExpenseCategoryDoc: %v\n", docSnap.Ref.ID, err)
			continue
		}
		category.ID = docSnap.Ref.ID // Populate ID.
		categories = append(categories, category)
	}

	// If the loop finishes without error and categories is empty, it means no active categories were found,
	// or the collection might be empty or non-existent (which Firestore handles gracefully for reads).
	return categories, nil
}
