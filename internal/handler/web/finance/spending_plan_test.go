package web_finance_test // Changed package name

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time" // For setting CreatedAt/UpdatedAt in test entities if needed

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	service_finance "github.com/Tomelin/dashfin-backend-app/internal/core/service/finance"
	web_finance "github.com/Tomelin/dashfin-backend-app/internal/handler/web/finance" // Changed import
)

// MockSpendingPlanService mocks the SpendingPlanService
type MockSpendingPlanService struct {
	mock.Mock
}

var _ service_finance.SpendingPlanService = &MockSpendingPlanService{} // Compile-time check

func (m *MockSpendingPlanService) GetSpendingPlan(ctx context.Context, userID string) (*entity_finance.SpendingPlan, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity_finance.SpendingPlan), args.Error(1)
}

func (m *MockSpendingPlanService) SaveSpendingPlan(ctx context.Context, userID string, planData *entity_finance.SpendingPlan) (*entity_finance.SpendingPlan, error) {
	args := m.Called(ctx, userID, planData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity_finance.SpendingPlan), args.Error(1)
}

// setupRouterAndMocks initializes a Gin router and the mock service for testing.
func setupRouterAndMocks() (*gin.Engine, *MockSpendingPlanService) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockService := new(MockSpendingPlanService)

	// Initialize the real handler with the mocked service
	// The web_finance.getUserIDFromContext will be called by the handler.
	// It's currently hardcoded to "temp-user-id-from-handler".
	handler := web_finance.NewSpendingPlanHandler(mockService) // Changed package

	// Setup routes similar to how they would be in main.go
	// Example: /api/v1/finance/spending-plan
	// For test purposes, we use /api/finance as the group.
	apiGroup := router.Group("/api/finance")
	handler.RegisterSpendingPlanRoutes(apiGroup) // This registers /spending-plan under apiGroup

	return router, mockService
}

// Test Cases will be implemented below
// TestGetSpendingPlan_Success
// TestGetSpendingPlan_NotFound
// TestGetSpendingPlan_ServiceError
// TestSaveSpendingPlan_Success
// TestSaveSpendingPlan_BadRequest_InvalidJSON
// TestSaveSpendingPlan_ServiceError
// TestSaveSpendingPlan_AuthError (If getUserIDFromContext can return error)

func TestGetSpendingPlan_Success(t *testing.T) {
	router, mockService := setupRouterAndMocks()

	userID := "temp-user-id-from-handler" // Must match what getUserIDFromContext returns
	expectedPlan := &entity_finance.SpendingPlan{
		UserID:        userID,
		MonthlyIncome: 5000,
		CategoryBudgets: []entity_finance.CategoryBudget{
			{Category: "Food", Amount: 500, Percentage: 0.1},
		},
		CreatedAt: time.Now().Add(-time.Hour), // Example time
		UpdatedAt: time.Now(),                 // Example time
	}

	mockService.On("GetSpendingPlan", mock.Anything, userID).Return(expectedPlan, nil).Once()

	req, _ := http.NewRequest(http.MethodGet, "/api/finance/spending-plan", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responsePlan entity_finance.SpendingPlan
	err := json.Unmarshal(w.Body.Bytes(), &responsePlan)
	assert.NoError(t, err)

	// Compare fields that are expected to be returned
	assert.Equal(t, expectedPlan.UserID, responsePlan.UserID)
	assert.Equal(t, expectedPlan.MonthlyIncome, responsePlan.MonthlyIncome)
	assert.Equal(t, expectedPlan.CategoryBudgets, responsePlan.CategoryBudgets)
	// For time fields, comparing UnixNano for exactness, assuming they are set and returned
	assert.Equal(t, expectedPlan.CreatedAt.UnixNano(), responsePlan.CreatedAt.UnixNano())
	assert.Equal(t, expectedPlan.UpdatedAt.UnixNano(), responsePlan.UpdatedAt.UnixNano())


	mockService.AssertExpectations(t)
}

func TestGetSpendingPlan_NotFound(t *testing.T) {
	router, mockService := setupRouterAndMocks()
	userID := "temp-user-id-from-handler"

	mockService.On("GetSpendingPlan", mock.Anything, userID).Return(nil, nil).Once()

	req, _ := http.NewRequest(http.MethodGet, "/api/finance/spending-plan", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	// Optionally, assert the error message in the response body
	// expectedError := `{"error":"Spending plan not found for this user"}`
	// assert.JSONEq(t, expectedError, w.Body.String())
	mockService.AssertExpectations(t)
}

func TestGetSpendingPlan_ServiceError(t *testing.T) {
	router, mockService := setupRouterAndMocks()
	userID := "temp-user-id-from-handler"
	serviceErr := errors.New("service failure")

	mockService.On("GetSpendingPlan", mock.Anything, userID).Return(nil, serviceErr).Once()

	req, _ := http.NewRequest(http.MethodGet, "/api/finance/spending-plan", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	// expectedError := `{"error":"Failed to retrieve spending plan","details":"service failure"}`
	// assert.JSONEq(t, expectedError, w.Body.String())
	mockService.AssertExpectations(t)
}

func TestSaveSpendingPlan_Success(t *testing.T) {
	router, mockService := setupRouterAndMocks()
	userID := "temp-user-id-from-handler"

	inputPlanData := entity_finance.SpendingPlan{
		MonthlyIncome: 6000,
		CategoryBudgets: []entity_finance.CategoryBudget{
			{Category: "Food", Amount: 100, Percentage: 0.016}, // Example
		},
	}
	// The service is expected to populate UserID, CreatedAt, UpdatedAt
	// For the purpose of matching the return, we create what we expect the service to return.
	expectedSavedPlan := inputPlanData
	expectedSavedPlan.UserID = userID
	// Actual CreatedAt/UpdatedAt would be set by service, so we can't predict exact values here
	// unless the service mock for SaveSpendingPlan returns specific time.
	// For now, we'll check UserID, MonthlyIncome, and CategoryBudgets in the response.
	// The mock should return a plan that includes these, plus any times it would set.

	// For the mock, we expect a plan that matches the input from the handler
	mockService.On("SaveSpendingPlan",
		mock.Anything, // Changed for context
		userID,
		mock.MatchedBy(func(argPlan *entity_finance.SpendingPlan) bool {
			return argPlan.MonthlyIncome == inputPlanData.MonthlyIncome &&
				assert.ObjectsAreEqual(inputPlanData.CategoryBudgets, argPlan.CategoryBudgets) &&
				argPlan.UserID == "" // UserID in planData passed to service might be empty or from request, service overrides
		}),
	).Run(func(args mock.Arguments) {
		// Simulate service setting UserID and time fields on the returned object
		// The actual object returned by the mock needs to be constructed carefully.
		// The mock will return a pointer to a SpendingPlan. Let's define it.
		// planArg := args.Get(2).(*entity_finance.SpendingPlan) // This was unused. The important part is what the mock returns.

		// What the mock call to SaveSpendingPlan should return:
		// It should be a *pointer* to a SpendingPlan struct.
		// This returned plan will have UserID, CreatedAt, UpdatedAt set by the service.
		// The mock setup's .Return() needs to provide this.
		// Let's make 'expectedSavedPlan' the one returned by the mock.
		// We need to ensure its UserID is set. CreatedAt/UpdatedAt will be different.
		// The important part for the mock is that the *input* to SaveSpendingPlan matches.
		// The *output* from SaveSpendingPlan (the first return arg of the mock) is what the handler gets back.

		// The mock should return a plan that looks like it came from the service
		// (i.e., UserID, CreatedAt, UpdatedAt are filled)
		// For this test, we'll make the mock return a version of planArg with UserID set.
		// To make assertions on the response easier, this returned object should be predictable.

		// The `expectedSavedPlan` variable already has UserID. Let's add plausible time.
		expectedSavedPlan.CreatedAt = time.Now().Add(-5 * time.Minute)
		expectedSavedPlan.UpdatedAt = time.Now()

	}).Return(&expectedSavedPlan, nil).Once()


	requestBodyBytes, _ := json.Marshal(inputPlanData)
	req, _ := http.NewRequest(http.MethodPut, "/api/finance/spending-plan", bytes.NewBuffer(requestBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responsePlan entity_finance.SpendingPlan
	err := json.Unmarshal(w.Body.Bytes(), &responsePlan)
	assert.NoError(t, err)

	// Assertions on the response from the handler
	assert.Equal(t, userID, responsePlan.UserID) // UserID should be set by service
	assert.Equal(t, inputPlanData.MonthlyIncome, responsePlan.MonthlyIncome)
	assert.Equal(t, inputPlanData.CategoryBudgets, responsePlan.CategoryBudgets)
	// For CreatedAt/UpdatedAt, we can check they are not zero if service sets them,
	// or compare to the specific times set in expectedSavedPlan if the mock returns it.
	assert.Equal(t, expectedSavedPlan.CreatedAt.UnixNano(), responsePlan.CreatedAt.UnixNano())
	assert.Equal(t, expectedSavedPlan.UpdatedAt.UnixNano(), responsePlan.UpdatedAt.UnixNano())


	mockService.AssertExpectations(t)
}


func TestSaveSpendingPlan_BadRequest_InvalidJSON(t *testing.T) {
	router, mockService := setupRouterAndMocks()

	req, _ := http.NewRequest(http.MethodPut, "/api/finance/spending-plan", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	// Service should not have been called
	mockService.AssertNotCalled(t, "SaveSpendingPlan", mock.Anything, mock.Anything, mock.Anything)
}

func TestSaveSpendingPlan_ServiceError(t *testing.T) {
	router, mockService := setupRouterAndMocks()
	userID := "temp-user-id-from-handler"

	inputPlanData := entity_finance.SpendingPlan{MonthlyIncome: 7000}
	serviceErr := errors.New("service save failure")

	mockService.On("SaveSpendingPlan",
		mock.Anything, // Changed for context
		userID,
		mock.MatchedBy(func(argPlan *entity_finance.SpendingPlan) bool {
			return argPlan.MonthlyIncome == inputPlanData.MonthlyIncome
		}),
	).Return(nil, serviceErr).Once()

	requestBodyBytes, _ := json.Marshal(inputPlanData)
	req, _ := http.NewRequest(http.MethodPut, "/api/finance/spending-plan", bytes.NewBuffer(requestBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
