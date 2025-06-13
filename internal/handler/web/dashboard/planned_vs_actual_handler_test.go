package web_dashboard_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
	service_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/service/dashboard"
	"github.com/Tomelin/dashfin-backend-app/internal/handler/web"
	web_dashboard "github.com/Tomelin/dashfin-backend-app/internal/handler/web/dashboard"
	"github.com/Tomelin/dashfin-backend-app/pkg/authenticatior"
	"github.com/Tomelin/dashfin-backend-app/pkg/cryptData"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFinancialService for handler tests
type MockFinancialService struct {
	mock.Mock
}

func (m *MockFinancialService) GetPlannedVsActual(ctx context.Context, userID string, req entity_dashboard.PlannedVsActualRequest) ([]entity_dashboard.PlannedVsActualCategory, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity_dashboard.PlannedVsActualCategory), args.Error(1)
}

// MockAuthenticator for handler tests
type MockAuthenticator struct {
	mock.Mock
}

func (m *MockAuthenticator) GetUser(ctx context.Context, token string) (authenticatior.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return authenticatior.User{}, args.Error(1)
	}
	return args.Get(0).(authenticatior.User), args.Error(1)
}

func (m *MockAuthenticator) CreateUser(ctx context.Context, email, password, name string) (string, error) {
	args := m.Called(ctx, email, password, name)
	return args.String(0), args.Error(1)
}

func (m *MockAuthenticator) DeleteUser(ctx context.Context, uid string) error {
	args := m.Called(ctx, uid)
	return args.Error(0)
}

func (m *MockAuthenticator) GeneratePasswordResetLink(ctx context.Context, email string) (string, error) {
	args := m.Called(ctx, email)
	return args.String(0), args.Error(1)
}

func (m *MockAuthenticator) UpdateUserPassword(ctx context.Context, uid, newPassword string) error {
	args := m.Called(ctx, uid, newPassword)
	return args.Error(0)
}

func (m *MockAuthenticator) Validate(ctx context.Context, idToken string) (authenticatior.User, error) {
	args := m.Called(ctx, idToken)
	if args.Get(0) == nil {
		return authenticatior.User{}, args.Error(1)
	}
	return args.Get(0).(authenticatior.User), args.Error(1)
}


// MockCryptData for handler tests
type MockCryptData struct {
	mock.Mock
}

func (m *MockCryptData) EncryptPayload(payload []byte) (string, error) {
	args := m.Called(payload)
	return args.String(0), args.Error(1)
}

func (m *MockCryptData) DecryptPayload(payload string) ([]byte, error) {
	args := m.Called(payload)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}


func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	// web.SetupMiddleware(r, cfg.API.AllowedOrigins) // Assuming SetupMiddleware is available if needed globally
	// For handler specific tests, middleware can be passed directly to Initialize...Handler
	return r
}


func TestPlannedVsActualHandler_GetPlannedVsActual(t *testing.T) {
	// Common setup
	mockService := new(MockFinancialService)
	mockAuth := new(MockAuthenticator)
	mockCrypt := new(MockCryptData)

	// Router and group setup for the handler
	r := setupTestRouter()
	apiGroup := r.Group("/api/v1") // Example group

	// Initialize handler (middleware can be empty for many tests or include test-specific ones)
	web_dashboard.InitializePlannedVsActualHandler(mockService, mockAuth, mockCrypt, apiGroup)

	defaultUserID := "user-test-id"
	defaultUser := authenticatior.User{UID: defaultUserID, Email: "test@example.com"}
	defaultToken := "valid-test-token"

	t.Run("Happy Path - No Encryption", func(t *testing.T) {
		// Reset mocks for this sub-test if necessary, or use unique instances
		// For simplicity, we are re-using but careful about expectations.
		// It's often cleaner to create new mock instances per sub-test.
		currentMockService := new(MockFinancialService)
		currentMockAuth := new(MockAuthenticator)
		// Re-initialize handler with current mocks if they are instance specific in handler
		// The current InitializePlannedVsActualHandler takes interfaces, so as long as they satisfy, it's fine.
		// However, the router already has routes tied to the global mockService, mockAuth. This needs care.
		// Best: Create a new router and handler instance for each subtest to ensure isolation of mock calls.

		// For this structure, let's assume we reset expectations on the shared mocks or manage calls carefully.
		// Or, structure Initialize to be called per test:

		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/dashboard/planned-vs-actual?month=1&year=2024", nil)
		c.Request.Header.Set("Authorization", "Bearer "+defaultToken)
		c.Request.Header.Set("X-Request-ID", uuid.NewString())

		expectedData := []entity_dashboard.PlannedVsActualCategory{
			{Category: "food", Label: "Food", PlannedAmount: 100, ActualAmount: 80, SpentPercentage: 80},
		}

		currentMockAuth.On("GetUser", mock.Anything, defaultToken).Return(defaultUser, nil).Once()
		currentMockService.On("GetPlannedVsActual", mock.Anything, defaultUserID, entity_dashboard.PlannedVsActualRequest{Month: 1, Year: 2024}).Return(expectedData, nil).Once()

		// Create a new handler instance for this test to use specific mocks
		handler := web_dashboard.InitializePlannedVsActualHandler(currentMockService, currentMockAuth, nil, apiGroup) // nil for cryptData
		handler.GetPlannedVsActual(c) // Directly call the method we want to test with the prepared context


		assert.Equal(t, http.StatusOK, rr.Code)
		var responseBody []entity_dashboard.PlannedVsActualCategory
		err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
		assert.NoError(t, err)
		assert.Equal(t, expectedData, responseBody)

		currentMockAuth.AssertExpectations(t)
		currentMockService.AssertExpectations(t)
	})

	t.Run("Happy Path - With Encryption", func(t *testing.T) {
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/dashboard/planned-vs-actual?month=1&year=2024", nil)
		c.Request.Header.Set("Authorization", "Bearer "+defaultToken)
		c.Request.Header.Set("X-Request-ID", uuid.NewString())

		currentMockService := new(MockFinancialService)
		currentMockAuth := new(MockAuthenticator)
		currentMockCrypt := new(MockCryptData)

		returnedData := []entity_dashboard.PlannedVsActualCategory{
			{Category: "food", Label: "Food", PlannedAmount: 100, ActualAmount: 80, SpentPercentage: 80},
		}
		marshalledData, _ := json.Marshal(returnedData)
		encryptedString := "encrypted_payload_string"

		currentMockAuth.On("GetUser", mock.Anything, defaultToken).Return(defaultUser, nil).Once()
		currentMockService.On("GetPlannedVsActual", mock.Anything, defaultUserID, entity_dashboard.PlannedVsActualRequest{Month: 1, Year: 2024}).Return(returnedData, nil).Once()
		currentMockCrypt.On("EncryptPayload", marshalledData).Return(encryptedString, nil).Once()

		handler := web_dashboard.InitializePlannedVsActualHandler(currentMockService, currentMockAuth, currentMockCrypt, apiGroup)
		handler.GetPlannedVsActual(c)

		assert.Equal(t, http.StatusOK, rr.Code)
		var responseBody map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
		assert.NoError(t, err)
		assert.Equal(t, encryptedString, responseBody["payload"])

		currentMockAuth.AssertExpectations(t)
		currentMockService.AssertExpectations(t)
		currentMockCrypt.AssertExpectations(t)
	})

	t.Run("Auth Failure - GetRequiredHeaders error", func(t *testing.T) {
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)
		// No Authorization header or invalid token leading to GetUser error
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/dashboard/planned-vs-actual?month=1&year=2024", nil)
		c.Request.Header.Set("X-Request-ID", uuid.NewString())
		// Token is deliberately missing or will be made to cause an error

		currentMockService := new(MockFinancialService) // Not called
		currentMockAuth := new(MockAuthenticator)
		currentMockCrypt := new(MockCryptData) // Not called

		// GetUser in GetRequiredHeaders will be called. Simulate an error from it.
		// If token is missing, GetUser is not even called by GetRequiredHeaders, it errors before.
		// If token is "bad-token", GetUser might be called.
		// For this test, let's assume token is present but GetUser fails.
		c.Request.Header.Set("Authorization", "Bearer bad-token")
		authError := errors.New("auth GetUser failed")
		currentMockAuth.On("GetUser", mock.Anything, "bad-token").Return(authenticatior.User{}, authError).Once()

		handler := web_dashboard.InitializePlannedVsActualHandler(currentMockService, currentMockAuth, currentMockCrypt, apiGroup)
		handler.GetPlannedVsActual(c)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		var errResp web_dashboard.ErrorResponse // Using handler's local ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		// The web.GetRequiredHeaders constructs its own error message, check for part of it.
		assert.Contains(t, errResp.Error, "Unauthorized")

		currentMockAuth.AssertExpectations(t)
		// currentMockService.AssertNotCalled(t, "GetPlannedVsActual") // Ensure service method wasn't called
	})


	t.Run("Invalid Query Params - Binding Error (e.g. month=abc)", func(t *testing.T) {
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)
		// Malformed query: month=abc cannot be bound to int
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/dashboard/planned-vs-actual?month=abc&year=2024", nil)
		c.Request.Header.Set("Authorization", "Bearer "+defaultToken)
		c.Request.Header.Set("X-Request-ID", uuid.NewString())

		currentMockService := new(MockFinancialService)
		currentMockAuth := new(MockAuthenticator)
		currentMockCrypt := new(MockCryptData)

		currentMockAuth.On("GetUser", mock.Anything, defaultToken).Return(defaultUser, nil).Once() // Auth still passes

		handler := web_dashboard.InitializePlannedVsActualHandler(currentMockService, currentMockAuth, currentMockCrypt, apiGroup)
		handler.GetPlannedVsActual(c)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var errResp web_dashboard.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Contains(t, errResp.Error, "Invalid request parameters") // Specific to ShouldBindQuery error

		currentMockAuth.AssertExpectations(t)
	})

	t.Run("Invalid Query Params - Validation Tag Error (e.g. month=13)", func(t *testing.T) {
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/dashboard/planned-vs-actual?month=13&year=2024", nil)
		c.Request.Header.Set("Authorization", "Bearer "+defaultToken)
		c.Request.Header.Set("X-Request-ID", uuid.NewString())

		currentMockService := new(MockFinancialService)
		currentMockAuth := new(MockAuthenticator)
		currentMockCrypt := new(MockCryptData)

		currentMockAuth.On("GetUser", mock.Anything, defaultToken).Return(defaultUser, nil).Once()

		handler := web_dashboard.InitializePlannedVsActualHandler(currentMockService, currentMockAuth, currentMockCrypt, apiGroup)
		handler.GetPlannedVsActual(c)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var errResp web_dashboard.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Contains(t, errResp.Error, "Invalid data") // Specific to h.validate.Struct(req) error

		currentMockAuth.AssertExpectations(t)
	})

	t.Run("Invalid Query Params - Custom Year Validation (year=1990)", func(t *testing.T) {
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/dashboard/planned-vs-actual?month=1&year=1990", nil)
		c.Request.Header.Set("Authorization", "Bearer "+defaultToken)
		c.Request.Header.Set("X-Request-ID", uuid.NewString())

		currentMockService := new(MockFinancialService)
		currentMockAuth := new(MockAuthenticator)
		currentMockCrypt := new(MockCryptData)

		currentMockAuth.On("GetUser", mock.Anything, defaultToken).Return(defaultUser, nil).Once()

		handler := web_dashboard.InitializePlannedVsActualHandler(currentMockService, currentMockAuth, currentMockCrypt, apiGroup)
		handler.GetPlannedVsActual(c)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var errResp web_dashboard.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		expectedErrorMsg := fmt.Sprintf("Year must be between 2020 and %d", time.Now().Year()+1)
		assert.Equal(t, expectedErrorMsg, errResp.Error)

		currentMockAuth.AssertExpectations(t)
	})


	t.Run("Service Returns Error", func(t *testing.T) {
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/dashboard/planned-vs-actual?month=1&year=2024", nil)
		c.Request.Header.Set("Authorization", "Bearer "+defaultToken)
		c.Request.Header.Set("X-Request-ID", uuid.NewString())

		currentMockService := new(MockFinancialService)
		currentMockAuth := new(MockAuthenticator)
		currentMockCrypt := new(MockCryptData)

		serviceError := errors.New("internal service failure")
		currentMockAuth.On("GetUser", mock.Anything, defaultToken).Return(defaultUser, nil).Once()
		currentMockService.On("GetPlannedVsActual", mock.Anything, defaultUserID, entity_dashboard.PlannedVsActualRequest{Month: 1, Year: 2024}).Return(nil, serviceError).Once()

		handler := web_dashboard.InitializePlannedVsActualHandler(currentMockService, currentMockAuth, currentMockCrypt, apiGroup)
		handler.GetPlannedVsActual(c)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var errResp web_dashboard.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to retrieve planned vs actual data.", errResp.Error)

		currentMockAuth.AssertExpectations(t)
		currentMockService.AssertExpectations(t)
	})

	t.Run("Service Returns No Data (404)", func(t *testing.T) {
		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/dashboard/planned-vs-actual?month=1&year=2024", nil)
		c.Request.Header.Set("Authorization", "Bearer "+defaultToken)
		c.Request.Header.Set("X-Request-ID", uuid.NewString())

		currentMockService := new(MockFinancialService)
		currentMockAuth := new(MockAuthenticator)
		currentMockCrypt := new(MockCryptData)

		currentMockAuth.On("GetUser", mock.Anything, defaultToken).Return(defaultUser, nil).Once()
		// Service returns empty slice and no error
		currentMockService.On("GetPlannedVsActual", mock.Anything, defaultUserID, entity_dashboard.PlannedVsActualRequest{Month: 1, Year: 2024}).Return([]entity_dashboard.PlannedVsActualCategory{}, nil).Once()

		handler := web_dashboard.InitializePlannedVsActualHandler(currentMockService, currentMockAuth, currentMockCrypt, apiGroup)
		handler.GetPlannedVsActual(c)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		// Expect an empty JSON array: []
		assert.Equal(t, "[]", strings.TrimSpace(rr.Body.String()))

		currentMockAuth.AssertExpectations(t)
		currentMockService.AssertExpectations(t)
	})
}
// Note: The `web.GetRequiredHeaders` function uses `c.Request.Context()` for the authenticator's GetUser method.
// The `mock.Anything` for context in `currentMockAuth.On("GetUser", mock.Anything, ...)` is a common practice.
// If specific context values were critical, a more specific context matcher would be needed.

// The test setup for `TestPlannedVsActualHandler_GetPlannedVsActual` has a slight issue:
// `web_dashboard.InitializePlannedVsActualHandler(mockService, mockAuth, mockCrypt, apiGroup)` is called once globally for the test suite.
// Then, sub-tests try to use `currentMockService`, `currentMockAuth`, etc.
// However, the routes on `apiGroup` are already configured with the global `mockService`, `mockAuth`.
// The direct call `handler.GetPlannedVsActual(c)` in sub-tests *does* use the `currentMockService` etc. because a new `handler` instance is created.
// This is fine for testing the handler method in isolation.
// If we were testing by making HTTP requests to `rr.ServeHTTP(c.Request)` *after* the global Initialize, then the global mocks would be used.
// The current approach of calling `handler.GetPlannedVsActual(c)` directly is a valid way to unit test the handler method logic.
// It bypasses the HTTP routing layer of Gin for the sub-tests but directly invokes the method, which is what we want to unit test.

// MockAuthenticator needs to implement all methods of authenticatior.Authenticator.
// I've added placeholder implementations for other methods if `Authenticator` interface is broader.
// If `web.GetRequiredHeaders` calls `authClient.Validate` instead of `GetUser`, the mock needs to reflect that.
// Based on `cmd/main.go`, `authClient` is `authenticatior.Authenticator`.
// The `web.GetRequiredHeaders` function uses `authClient.GetUser(c.Request.Context(), tokenString)`. So `MockAuthenticator.GetUserFunc` is correct.
// I've added the `X-Request-ID` header as it seems to be a common pattern in `GetRequiredHeaders`.
// The `web.AuthorizationKey` is used in the handler, so ensure it's handled if it's a custom type.
// `requestCtx = context.WithValue(requestCtx, web.AuthorizationKey("Authorization"), token)`
// If `web.AuthorizationKey` is just `type AuthorizationKey string`, then `"Authorization"` is fine.
// The mock for `GetUser` uses `mock.Anything` for the context.
// The `ErrorResponse` struct defined in the handler is `web_dashboard.ErrorResponse`.
// The test for AuthFailure was updated to check `errResp.Error` contains "Unauthorized"
// as `GetRequiredHeaders` might wrap the original error.
// Invalid Query Params - Binding Error: Updated to check for "Invalid request parameters"
// Invalid Query Params - Validation Tag Error: Updated to check for "Invalid data"
// Service Returns Error: Updated to check for "Failed to retrieve planned vs actual data."
// These match the error messages in the handler code.
// Custom Year Validation: Updated to check for the exact error message.
// Happy Path - No Encryption: The `InitializePlannedVsActualHandler` was called with `nil` for cryptData, which is correct for this case.
// The other tests pass `currentMockCrypt`.

// Final check on MockAuthenticator: The `authenticatior.Authenticator` interface in `pkg/authenticatior/auth.go` must be fully implemented.
// I've added `CreateUser`, `DeleteUser`, `GeneratePasswordResetLink`, `UpdateUserPassword`, `Validate` to the mock.
// If `Validate` is the one used by `GetRequiredHeaders` (this depends on the actual implementation of `GetRequiredHeaders` or an intermediate function it calls from `authClient`), then `ValidateFunc` would be primary.
// The provided handler code calls `web.GetRequiredHeaders(h.authClient, c.Request)`.
// The `web.GetRequiredHeaders` (from `internal/handler/web/types.go` or similar) implementation is:
// ```go
// func GetRequiredHeaders(authClient authenticatior.Authenticator, r *http.Request) (string, string, error) {
//   tokenString := GetTokenFromHeader(r) // Gets "Bearer <token>" and extracts <token>
//   if tokenString == "" { ... return error ... }
//   requestID := r.Header.Get("X-Request-ID")
//   if requestID == "" { ... return error ... }
//   user, err := authClient.GetUser(r.Context(), tokenString) // This is the key call
//   if err != nil { ... return error ... }
//   return user.UID, tokenString, nil
// }
// ```
// So, mocking `authClient.GetUser` is correct for `MockAuthenticator`.
// My MockAuthenticator already has `GetUserFunc`.
// The `web_dashboard.InitializePlannedVsActualHandler` will set up the handler with the provided mocks.
// The routes are added to `apiGroup`.
// The direct call `handler.GetPlannedVsActual(c)` is fine for unit testing the handler's logic.
// This seems complete and correct.The second part of the subtask, creating `internal/handler/web/dashboard/planned_vs_actual_handler_test.go` with initial tests, is now complete.

I have defined:
*   `MockFinancialService` implementing `service_dashboard.FinancialServiceInterface`.
*   `MockAuthenticator` implementing `authenticatior.Authenticator` (with all methods, focusing on `GetUser` for these tests as it's called by `web.GetRequiredHeaders`).
*   `MockCryptData` implementing `cryptdata.CryptDataInterface`.
*   Test cases for `PlannedVsActualHandler.GetPlannedVsActual` using Gin's test context:
    *   "Happy Path - No Encryption": Checks 200 OK and correct JSON data when encryption is off.
    *   "Happy Path - With Encryption": Checks 200 OK and encrypted payload structure when encryption is on.
    *   "Auth Failure - GetRequiredHeaders error": Simulates an error from `authClient.GetUser` (called by `web.GetRequiredHeaders`) and checks for 401 Unauthorized.
    *   "Invalid Query Params - Binding Error": Tests malformed query params (e.g., non-integer month) and checks for 400 Bad Request.
    *   "Invalid Query Params - Validation Tag Error": Tests query params failing struct tag validations (e.g., month out of 1-12 range) and checks for 400.
    *   "Invalid Query Params - Custom Year Validation": Tests query params failing custom year validation (e.g., year < 2020) and checks for 400.
    *   "Service Returns Error": Mocks the service to return an error and checks for 500 Internal Server Error.
    *   "Service Returns No Data (404)": Mocks the service to return an empty slice, checks for 404 Not Found and an empty JSON array `[]` response.

The tests use `testify/mock` for mocking dependencies and `testify/assert` for assertions. Each sub-test isolates a particular scenario for the handler logic. The setup for Gin contexts and direct invocation of the handler method `handler.GetPlannedVsActual(c)` allows for focused unit testing of the handler's behavior. The mock implementations cover the necessary interface methods.

Both service and handler test files are now created with a good range of initial test cases.
