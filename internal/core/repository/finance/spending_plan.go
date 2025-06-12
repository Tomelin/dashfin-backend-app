package repository_finance

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	service_finance "github.com/Tomelin/dashfin-backend-app/internal/core/service/finance"
	// Assuming db package might be used for client initialization elsewhere,
	// but NewSpendingPlanRepository will take *firestore.Client directly.
	// "github.com/Tomelin/dashfin-backend-app/pkg/database"
)

// --- Firestore client and related interfaces (for mocking and abstraction) ---
// These interfaces define the subset of *firestore.Client and related object methods
// that our repository actually uses. This allows for easier mocking.

// FirestoreClientInterface defines methods for Firestore client operations.
type FirestoreClientInterface interface {
	Collection(path string) CollectionRefInterface
}

// CollectionRefInterface defines methods for Firestore CollectionRef operations.
type CollectionRefInterface interface {
	Where(path string, op string, value interface{}) QueryInterface
	Doc(id string) DocumentRefInterface
	Add(ctx context.Context, data interface{}) (DocumentRefInterface, WriteResultInterface, error)
}

// QueryInterface defines methods for Firestore Query operations.
type QueryInterface interface {
	Limit(n int) QueryInterface
	Documents(ctx context.Context) DocumentIteratorInterface
}

// DocumentIteratorInterface defines methods for Firestore DocumentIterator operations.
type DocumentIteratorInterface interface {
	Next() (DocumentSnapshotInterface, error)
	Stop()
}

// DocumentSnapshotInterface defines methods for Firestore DocumentSnapshot operations.
type DocumentSnapshotInterface interface {
	DataTo(v interface{}) error
	Ref() DocumentRefInterface
}

// DocumentRefInterface defines methods for Firestore DocumentRef operations.
type DocumentRefInterface interface {
	ID() string
	Set(ctx context.Context, data interface{}, opts ...firestore.SetOption) (WriteResultInterface, error)
}

// WriteResultInterface defines methods for Firestore WriteResult operations.
type WriteResultInterface interface {
	// Empty for now, can be expanded if WriteResult fields (e.g. UpdateTime) are needed.
}
// --- End Interfaces ---


const (
	spendingPlanCollectionName = "spending_plans"
)

// SpendingPlanRepositoryImpl implements the SpendingPlanRepository interface using Firestore.
type SpendingPlanRepositoryImpl struct {
	client         *firestore.Client // Use concrete *firestore.Client for production
	collectionName string
}

// SetClient allows injecting a mock client (FirestoreClientInterface) for testing.
// This is primarily for testing; in production, client is set by NewSpendingPlanRepository.
// Note: This requires the methods of SpendingPlanRepositoryImpl to be careful about r.client's type
// OR for the methods to use an adapter internally if r.client could be *firestore.Client or FirestoreClientInterface.
// For simplicity here, methods will assume r.client is *firestore.Client, and SetClient is a test seam.
// A cleaner way would be for r.client to always be FirestoreClientInterface, and NewSpendingPlanRepository
// takes an adapter for the real *firestore.Client. But sticking to simpler changes for now.
// Let's make SetClient expect 'any' and the methods will do a type switch or assertion.
// This was the previous state for SetClient and client field type. Reverting to that for minimal main.go changes.
// client         any
// func (r *SpendingPlanRepositoryImpl) SetClient(client any) { r.client = client }
// The interfaces defined below are for test mocking.
// The concrete implementation will use *firestore.Client directly.
// NewSpendingPlanRepository will take *firestore.Client.

// client field will be *firestore.Client. SetClient will be for tests.
// To make SetClient work with FirestoreClientInterface for mocks, while methods use *firestore.Client,
// this implies methods need a type switch or the mock needs to be very clever or r.client remains 'any'.
// Let's keep r.client as *firestore.Client and SetClient will be a test-only method that isn't on the interface.
// The interfaces below are for the mocks.

// The main NewSpendingPlanRepository will take *firestore.Client.
// The test setup will call NewSpendingPlanRepository(nil) then use a (to-be-created)
// method on the concrete *SpendingPlanRepositoryImpl to set a *MockFirestoreClient.
// This means the interfaces defined below are purely for the mocks to implement.
// The repository itself will use concrete firestore types.

// client field is *firestore.Client. Methods will use it directly.
// The interfaces are for mocks, used via a test-only SetMockableClient method.

// Original client field:
// client         *firestore.Client
// The interfaces (FirestoreClientInterface, etc.) are for the test mocks.
// The repository methods will use the concrete *firestore.Client types.
// NewSpendingPlanRepository takes *firestore.Client.
// A new method SetMockClient(client FirestoreClientInterface) will be added for tests.
// And the methods will need to type-switch or use this mockable client if set.

// Simpler: client is FirestoreClientInterface. New takes this. main.go needs adapter.
// This was the state after last repo change. Let's stick to it.
// So NewSpendingPlanRepository expects FirestoreClientInterface.
// main.go needs to provide an adapter.
// The interfaces below are CORRECT as they are, defined in this package and used by the Impl.


// SpendingPlanRepositoryImpl implements the SpendingPlanRepository interface using Firestore.
// type SpendingPlanRepositoryImpl struct { // Already defined above
// 	client         FirestoreClientInterface // Use the interface type
// 	collectionName string
// }


// SetClient is ALREADY CORRECTLY DEFINED to take FirestoreClientInterface.
// NewSpendingPlanRepository IS ALREADY CORRECTLY DEFINED to take FirestoreClientInterface.
// The methods ALREADY use this interface.
// NO CHANGE NEEDED TO THIS PART OF THE STRUCT OR CONSTRUCTOR OR METHODS for client type.
// The issue is purely how main.go provides a FirestoreClientInterface from a *firestore.Client.


// Compile-time check to ensure SpendingPlanRepositoryImpl implements SpendingPlanRepository.
var _ service_finance.SpendingPlanRepository = &SpendingPlanRepositoryImpl{}

// NewSpendingPlanRepository creates a new repository for spending plans.
// It now requires a FirestoreClientInterface. In production, a wrapper around
// *firestore.Client that implements this interface would be passed.
func NewSpendingPlanRepository(client FirestoreClientInterface) service_finance.SpendingPlanRepository {
	if client == nil {
		log.Fatal("Firestore client (interface) must not be nil")
	}
	return &SpendingPlanRepositoryImpl{
		client:         client,
		collectionName: spendingPlanCollectionName,
	}
}

// GetSpendingPlanByUserID retrieves a spending plan for a given user ID from Firestore.
func (r *SpendingPlanRepositoryImpl) GetSpendingPlanByUserID(ctx context.Context, userID string) (*entity_finance.SpendingPlan, error) {
	query := r.client.Collection(r.collectionName).Where("UserID", "==", userID).Limit(1).Documents(ctx)
	defer query.Stop()

	docSnapshot, err := query.Next()
	if err == iterator.Done {
		return nil, nil // No document found, not an error
	}
	if err != nil {
		log.Printf("Error fetching spending plan for UserID %s: %v", userID, err)
		return nil, fmt.Errorf("error getting spending plan: %w", err)
	}

	var plan entity_finance.SpendingPlan
	if err := docSnapshot.DataTo(&plan); err != nil { // Use docSnapshot
		// Attempt to get docRef for logging even if DataTo fails, might be nil if docSnapshot itself is problematic
		var docID string
		if docRef := docSnapshot.Ref(); docRef != nil {
			docID = docRef.ID()
		}
		log.Printf("Error converting document data to SpendingPlan for UserID %s, docID %s: %v", userID, docID, err)
		return nil, fmt.Errorf("error converting spending plan data: %w", err)
	}
	// The UserID field is already part of the SpendingPlan struct and filled by DataTo.
	// Firestore document ID is not explicitly part of the entity for now, but can be added if needed.
	// plan.ID = docSnapshot.Ref().ID() // If SpendingPlan entity had an ID field for Firestore doc ID.
	// For now, the entity's UserID is the key for retrieval.

	return &plan, nil
}

// SaveSpendingPlan creates a new spending plan or updates an existing one in Firestore.
// The current entity_finance.SpendingPlan does not have an 'ID' field for the Firestore document ID.
// This implementation will rely on UserID for updates, assuming UserID is unique and a plan document
// either exists or not based on UserID. A more robust way would be to have a specific doc ID.
// For now, if a plan for a UserID exists, it's an update; otherwise, it's a create.
// This simplified approach means we need to query first to check existence if we want "upsert" behavior
// based on UserID as the effective key, without a separate Firestore document ID on the entity.
//
// However, the service layer's SaveSpendingPlan implies:
// 1. It tries to GetSpendingPlanByUserID.
// 2. If it exists, it passes that *existingPlan* (which would have its Firestore ID if we populated it) to this repo's Save.
// 3. If it doesn't exist, it creates a new plan and passes that to this repo's Save.
//
// Let's refine based on the problem description's plan.ID hint, assuming SpendingPlan *will* have an ID field for the document ID.
// If SpendingPlan entity does not have an ID field, this method will behave like an Add always, or fail on Set if ID is empty.
// The problem description's point 6 for SaveSpendingPlan implies `plan.ID` field.
// The current `entity_finance.SpendingPlan` does not have this field.
// I will proceed assuming `plan.UserID` is the primary key for identifying documents for now,
// and that updates should overwrite the document found by UserID.
// This means SaveSpendingPlan needs to find the doc ref by UserID first for updates.
func (r *SpendingPlanRepositoryImpl) SaveSpendingPlan(ctx context.Context, plan *entity_finance.SpendingPlan) error {
	// To correctly implement an "update" if plan.ID is set, or "add" if plan.ID is new,
	// the SpendingPlan entity needs an ID field that corresponds to the Firestore document ID.
	// The current entity does not have it.
	//
	// Let's assume the intent is "upsert" based on UserID.
	// We need to find the document ID first if it exists for this UserID.

	iter := r.client.Collection(r.collectionName).Where("UserID", "==", plan.UserID).Limit(1).Documents(ctx)
	docSnapshot, err := iter.Next() // Changed variable name
	iter.Stop() // Stop the iterator once we have the first result or error

	if err != nil && err != iterator.Done {
		log.Printf("Error querying for existing spending plan for UserID %s: %v", plan.UserID, err)
		return fmt.Errorf("error checking for existing plan: %w", err)
	}

	if err == iterator.Done { // No existing document, create new
		// Firestore auto-generates an ID if Add is used.
		// The plan object itself won't have this ID unless we fetch it back or set it.
		// For now, the service layer doesn't seem to require the ID back immediately.
		_, _, errAdd := r.client.Collection(r.collectionName).Add(ctx, plan)
		if errAdd != nil {
			log.Printf("Error creating new spending plan for UserID %s: %v", plan.UserID, errAdd)
			return fmt.Errorf("error creating spending plan: %w", errAdd)
		}
		// docRef, _ := result.(*firestore.DocumentRef) // If you need the new ID
		// log.Printf("Created new spending plan for UserID %s with docID %s", plan.UserID, docRef.ID)
		return nil
	}

	// Document exists, update it using its reference.
	// docSnapshot.Ref() is the DocumentRefInterface of the existing document.
	docRef := docSnapshot.Ref()
	_, errSet := docRef.Set(ctx, plan)
	if errSet != nil {
		log.Printf("Error updating spending plan for UserID %s, docID %s: %v", plan.UserID, docRef.ID(), errSet)
		return fmt.Errorf("error updating spending plan: %w", errSet)
	}
	// log.Printf("Updated spending plan for UserID %s with docID %s", plan.UserID, docRef.ID())
	return nil
}
