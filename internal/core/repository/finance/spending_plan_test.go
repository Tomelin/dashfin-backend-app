package repository_finance_test

import (
	"context"
	// "errors" // Removed
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	repository_finance "github.com/Tomelin/dashfin-backend-app/internal/core/repository/finance"
	service_finance "github.com/Tomelin/dashfin-backend-app/internal/core/service/finance"
)

// Mock Objects Start

// MockFirestoreClient mocks repository_finance.FirestoreClientInterface
type MockFirestoreClient struct {
	mock.Mock
}
var _ repository_finance.FirestoreClientInterface = &MockFirestoreClient{} // Compile-time check

func (m *MockFirestoreClient) Collection(path string) repository_finance.CollectionRefInterface {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(repository_finance.CollectionRefInterface)
}

// MockCollectionRef mocks repository_finance.CollectionRefInterface
type MockCollectionRef struct {
	mock.Mock
}
var _ repository_finance.CollectionRefInterface = &MockCollectionRef{} // Compile-time check

func (m *MockCollectionRef) Where(path string, op string, value interface{}) repository_finance.QueryInterface {
	args := m.Called(path, op, value)
	return args.Get(0).(repository_finance.QueryInterface)
}

func (m *MockCollectionRef) Doc(id string) repository_finance.DocumentRefInterface {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(repository_finance.DocumentRefInterface)
}

func (m *MockCollectionRef) Add(ctx context.Context, data interface{}) (repository_finance.DocumentRefInterface, repository_finance.WriteResultInterface, error) {
	args := m.Called(ctx, data)
	r0, _ := args.Get(0).(repository_finance.DocumentRefInterface)
	r1, _ := args.Get(1).(repository_finance.WriteResultInterface)
	return r0, r1, args.Error(2)
}

// MockQuery mocks repository_finance.QueryInterface
type MockQuery struct {
	mock.Mock
}
var _ repository_finance.QueryInterface = &MockQuery{} // Compile-time check

func (m *MockQuery) Limit(n int) repository_finance.QueryInterface {
	args := m.Called(n)
	return args.Get(0).(repository_finance.QueryInterface)
}

func (m *MockQuery) Documents(ctx context.Context) repository_finance.DocumentIteratorInterface {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(repository_finance.DocumentIteratorInterface)
}

// MockDocumentIterator mocks repository_finance.DocumentIteratorInterface
type MockDocumentIterator struct {
	mock.Mock
}
var _ repository_finance.DocumentIteratorInterface = &MockDocumentIterator{} // Compile-time check

func (m *MockDocumentIterator) Next() (repository_finance.DocumentSnapshotInterface, error) {
	args := m.Called()
	r0, _ := args.Get(0).(repository_finance.DocumentSnapshotInterface)
	return r0, args.Error(1)
}

func (m *MockDocumentIterator) Stop() {
	m.Called()
}

// MockDocumentSnapshot mocks repository_finance.DocumentSnapshotInterface
type MockDocumentSnapshot struct {
	mock.Mock
}
var _ repository_finance.DocumentSnapshotInterface = &MockDocumentSnapshot{} // Compile-time check

func (m *MockDocumentSnapshot) DataTo(v interface{}) error {
	args := m.Called(v)
	// Simulate filling the struct by using mock.Run in the test expectation for DataTo
	// Example: mockSnap.On("DataTo", mock.AnythingOfType("*entity_finance.SpendingPlan")).Run(func(args mock.Arguments) { ... })
	return args.Error(0)
}


func (m *MockDocumentSnapshot) Ref() repository_finance.DocumentRefInterface {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(repository_finance.DocumentRefInterface)
}

// MockDocumentRef mocks repository_finance.DocumentRefInterface
type MockDocumentRef struct {
	mock.Mock
}
var _ repository_finance.DocumentRefInterface = &MockDocumentRef{} // Compile-time check

func (m *MockDocumentRef) ID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDocumentRef) Set(ctx context.Context, data interface{}, opts ...firestore.SetOption) (repository_finance.WriteResultInterface, error) {
	args := m.Called(ctx, data) // Ignoring opts for now
	r0, _ := args.Get(0).(repository_finance.WriteResultInterface)
	return r0, args.Error(1)
}

// MockWriteResult mocks repository_finance.WriteResultInterface
type MockWriteResult struct {
	mock.Mock
}
var _ repository_finance.WriteResultInterface = &MockWriteResult{} // Compile-time check


// Mock Objects End

// Helper function to setup repository with mocks
func setupRepoWithMocks(t *testing.T) (service_finance.SpendingPlanRepository, *MockFirestoreClient) {
	mockClient := new(MockFirestoreClient)
	// NewSpendingPlanRepository now accepts FirestoreClientInterface
	repo := repository_finance.NewSpendingPlanRepository(mockClient)
	return repo, mockClient
}


// TestGetSpendingPlanByUserID_Found (Implementation in next step)

func TestGetSpendingPlanByUserID_Found(t *testing.T) {
	repo, mockClient := setupRepoWithMocks(t)
	ctx := context.Background()
	userID := "user123"
	expectedPlan := &entity_finance.SpendingPlan{
		UserID:        userID,
		MonthlyIncome: 5000,
		CategoryBudgets: []entity_finance.CategoryBudget{
			{Category: "Food", Amount: 500, Percentage: 0.1},
		},
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
	}

	// Mock setup
	mockCollection := new(MockCollectionRef)
	mockQuery := new(MockQuery)
	mockIter := new(MockDocumentIterator)
	mockSnap := new(MockDocumentSnapshot)
	// MockDocumentRef is not strictly needed for Get operation unless we use its ID.
	// mockDocRef := new(MockDocumentRef)

	mockClient.On("Collection", "spending_plans").Return(mockCollection)
	mockCollection.On("Where", "UserID", "==", userID).Return(mockQuery)
	mockQuery.On("Limit", 1).Return(mockQuery)
	mockQuery.On("Documents", ctx).Return(mockIter)

	// For "Found" case with Limit(1), Next is called once successfully.
	// The second call to Next that would yield iterator.Done is not made.
	mockIter.On("Next").Return(mockSnap, nil).Once()
	mockIter.On("Stop").Return(nil).Once()

	mockSnap.On("DataTo", mock.AnythingOfType("*entity_finance.SpendingPlan")).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*entity_finance.SpendingPlan)
		*arg = *expectedPlan // Populate the passed-in struct
	}).Return(nil)
	// If plan.ID = doc.Ref.ID was used in repo:
	// mockSnap.On("Ref").Return(mockDocRef)
	// mockDocRef.On("ID").Return("test-doc-id")


	plan, err := repo.GetSpendingPlanByUserID(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, expectedPlan.UserID, plan.UserID)
	assert.Equal(t, expectedPlan.MonthlyIncome, plan.MonthlyIncome)
	assert.Equal(t, expectedPlan.CategoryBudgets, plan.CategoryBudgets)
	// Compare time with some tolerance or by UnixNano for exact match if clocks are identical
	assert.Equal(t, expectedPlan.CreatedAt.UnixNano(), plan.CreatedAt.UnixNano())
	assert.Equal(t, expectedPlan.UpdatedAt.UnixNano(), plan.UpdatedAt.UnixNano())

	mockClient.AssertExpectations(t)
	mockCollection.AssertExpectations(t)
	mockQuery.AssertExpectations(t)
	mockIter.AssertExpectations(t)
	mockSnap.AssertExpectations(t)
	// mockDocRef.AssertExpectations(t) // If used
}

func TestGetSpendingPlanByUserID_NotFound(t *testing.T) {
	repo, mockClient := setupRepoWithMocks(t)
	ctx := context.Background()
	userID := "user404"

	// Mock setup
	mockCollection := new(MockCollectionRef)
	mockQuery := new(MockQuery)
	mockIter := new(MockDocumentIterator)

	mockClient.On("Collection", "spending_plans").Return(mockCollection)
	mockCollection.On("Where", "UserID", "==", userID).Return(mockQuery)
	mockQuery.On("Limit", 1).Return(mockQuery)
	mockQuery.On("Documents", ctx).Return(mockIter)

	mockIter.On("Next").Return(nil, iterator.Done).Once() // Simulate not found
	mockIter.On("Stop").Return(nil).Once()

	plan, err := repo.GetSpendingPlanByUserID(ctx, userID)

	assert.NoError(t, err)
	assert.Nil(t, plan)

	mockClient.AssertExpectations(t)
	mockCollection.AssertExpectations(t)
	mockQuery.AssertExpectations(t)
	mockIter.AssertExpectations(t)
}

func TestSaveSpendingPlan_CreateNew(t *testing.T) {
	repo, mockClient := setupRepoWithMocks(t)
	ctx := context.Background()
	newPlan := &entity_finance.SpendingPlan{
		UserID:        "newUser123",
		MonthlyIncome: 3000,
		CategoryBudgets: []entity_finance.CategoryBudget{
			{Category: "Savings", Amount: 300, Percentage: 0.1},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock setup for the initial query (to find if doc exists)
	mockCollectionQuery := new(MockCollectionRef)
	mockQuery := new(MockQuery)
	mockIter := new(MockDocumentIterator)

	// This client.Collection call is for the initial query in SaveSpendingPlan
	mockClient.On("Collection", "spending_plans").Return(mockCollectionQuery).Once() // Expect this call once for the query
	mockCollectionQuery.On("Where", "UserID", "==", newPlan.UserID).Return(mockQuery)
	mockQuery.On("Limit", 1).Return(mockQuery)
	mockQuery.On("Documents", ctx).Return(mockIter)
	mockIter.On("Next").Return(nil, iterator.Done).Once() // Simulate not found
	mockIter.On("Stop").Return(nil).Once()

	// Mock setup for the Add operation
	mockCollectionAdd := new(MockCollectionRef) // Separate mock instance for clarity if needed, or reuse if appropriate
	mockDocRef := new(MockDocumentRef)
	mockWriteResult := new(MockWriteResult)

	// This client.Collection call is for the Add operation in SaveSpendingPlan
	// If the same mockCollectionQuery instance is expected by the client for both, adjust accordingly.
	// For distinctness in test, or if different CollectionRef objects are used:
	mockClient.On("Collection", "spending_plans").Return(mockCollectionAdd).Once() // Expect this call once for the Add
	mockCollectionAdd.On("Add", ctx, newPlan).Return(mockDocRef, mockWriteResult, nil)
	// We don't use mockDocRef.ID or mockWriteResult in the current repo.SaveSpendingPlan, so no On calls for them are needed.

	err := repo.SaveSpendingPlan(ctx, newPlan)

	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
	mockCollectionQuery.AssertExpectations(t)
	mockQuery.AssertExpectations(t)
	mockIter.AssertExpectations(t)
	mockCollectionAdd.AssertExpectations(t)
	// mockDocRef.AssertExpectations(t) // Not called upon by current repo logic after Add
	// mockWriteResult.AssertExpectations(t) // Not used
}

func TestSaveSpendingPlan_UpdateExisting(t *testing.T) {
	repo, mockClient := setupRepoWithMocks(t)
	ctx := context.Background()
	existingPlan := &entity_finance.SpendingPlan{
		UserID:        "existingUser123",
		MonthlyIncome: 7000,
		CategoryBudgets: []entity_finance.CategoryBudget{
			{Category: "Travel", Amount: 700, Percentage: 0.1},
		},
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-30 * time.Minute), // Older UpdatedAt
	}
	// updatedPlanData := *existingPlan // Create a copy to modify for update
	// updatedPlanData.MonthlyIncome = 7500
	// updatedPlanData.UpdatedAt = time.Now() // Service layer would set this

	// Mock setup for the initial query (to find the existing doc)
	mockCollectionQuery := new(MockCollectionRef)
	mockQuery := new(MockQuery)
	mockIter := new(MockDocumentIterator)
	mockExistingSnap := new(MockDocumentSnapshot)
	mockExistingDocRef := new(MockDocumentRef) // This will be the Ref of the found document

	// This client.Collection call is for the initial query
	mockClient.On("Collection", "spending_plans").Return(mockCollectionQuery).Once()
	mockCollectionQuery.On("Where", "UserID", "==", existingPlan.UserID).Return(mockQuery)
	mockQuery.On("Limit", 1).Return(mockQuery)
	mockQuery.On("Documents", ctx).Return(mockIter)

	mockIter.On("Next").Return(mockExistingSnap, nil).Once() // Simulate doc found
	mockIter.On("Stop").Return(nil).Once()                 // Stop after finding the doc

	// Mock the found snapshot and its reference
	// DataTo is not strictly needed for Save if the repo doesn't re-read data, but Ref is crucial.
	// mockExistingSnap.On("DataTo", mock.AnythingOfType("*entity_finance.SpendingPlan")).Return(nil) // Not strictly needed for Save
	mockExistingSnap.On("Ref").Return(mockExistingDocRef)
	// The ID of mockExistingDocRef is not explicitly used by repo.Save if it just uses the Ref.

	// Mock setup for the Set operation
	mockWriteResult := new(MockWriteResult)
	// The existingPlan object (or a modified version of it) will be passed to Set.
	// The repository's SaveSpendingPlan uses the doc.Ref from the query to do the Set.
	// So, mockExistingDocRef is the one on which Set will be called.
	mockExistingDocRef.On("Set", ctx, existingPlan).Return(mockWriteResult, nil)


	err := repo.SaveSpendingPlan(ctx, existingPlan) // Pass existingPlan, which is what the service would update and pass

	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
	mockCollectionQuery.AssertExpectations(t)
	mockQuery.AssertExpectations(t)
	mockIter.AssertExpectations(t)
	mockExistingSnap.AssertExpectations(t)
	mockExistingDocRef.AssertExpectations(t)
	// mockWriteResult.AssertExpectations(t) // Not used
}


// Placeholder for tests to make the file compile
func TestPlaceholder(t *testing.T) {
	// Example of how setupRepoWithMocks would be used:
	// repo, mockClient := setupRepoWithMocks(t)
	// assert.NotNil(t, repo)
	// assert.NotNil(t, mockClient)
	assert.True(t, true)
}

// NOTE: The repository's 'client' field was changed to 'any' and a SetClient method added.
// The repository methods now perform a type assertion on this 'client' field.
// This setup is a workaround for testing with testify/mock when the original code
// depends on a concrete type (*firestore.Client) rather than an interface.
