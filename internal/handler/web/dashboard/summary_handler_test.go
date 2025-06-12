// Package dashboard_web_test contains tests for the dashboard HTTP handler.
package dashboard_web_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
	service_dashboard_mocks "github.com/Tomelin/dashfin-backend-app/internal/core/service/dashboard/mocks"
	dashboard_web "github.com/Tomelin/dashfin-backend-app/internal/handler/web/dashboard"
	"github.com/Tomelin/dashfin-backend-app/pkg/authenticatior" // Assuming a mock or dummy can be made if needed
	"github.com/Tomelin/dashfin-backend-app/pkg/cryptData"      // Assuming a mock or dummy
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthMiddleware is a dummy auth middleware for testing.
// It simulates setting the user_id in the context.
func MockAuthMiddleware(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if userID != "" {
			c.Set("user_id", userID)
		}
		c.Next()
	}
}

// MockHeaderMiddleware is a dummy middleware.
func MockHeaderMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// MockAuthenticator is a simple mock for authenticatior.Authenticator
type MockAuthenticator struct {
	mock.Mock
	UserIDToSet string // control what userID this mock will set
}

func (m *MockAuthenticator) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.UserIDToSet != "" {
			c.Set("user_id", m.UserIDToSet)
		}
		// Simulate token validation logic if necessary for more complex tests
		// For now, just sets userID if UserIDToSet is configured.
		c.Next()
	}
}
func (m *MockAuthenticator) GenerateToken(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}
func (m *MockAuthenticator) ValidateToken(ctx context.Context, token string) (authenticatior.AuthUser, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(authenticatior.AuthUser), args.Error(1)
}


// MockCryptData is a simple mock for cryptdata.CryptDataInterface
type MockCryptData struct {mock.Mock}
func (m *MockCryptData) Encrypt(data string) (string, error) { args := m.Called(data); return args.String(0), args.Error(1) }
func (m *MockCryptData) Decrypt(data string) (string, error) { args := m.Called(data); return args.String(0), args.Error(1) }


func TestDashboardHandler_GetDashboardSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(service_dashboard_mocks.MockDashboardService)

	// Create mock authenticator and cryptData
	// These are needed by NewDashboardHandler but might not be deeply involved in GET /summary logic itself
	// beyond authClient.Middleware() setting the userID.
	mockAuth := &MockAuthenticator{}
	mockCrypt := &MockCryptData{}


	// Setup router and handler
	// The handler constructor takes a router group.
	// We create a test router and a group to pass.
	router := gin.New()
	apiGroup := router.Group("/api") // Assuming handlers are typically grouped

	// The handler registers its routes upon creation.
	// No need to pass corsMiddleware if it's handled globally or not specifically tested here.
	// The MockHeaderMiddleware is a placeholder.
	dashboard_web.NewDashboardHandler(mockService, mockCrypt, mockAuth, apiGroup, nil, MockHeaderMiddleware())


	t.Run("Successful Request", func(t *testing.T) {
		// Reset mocks for sub-test if necessary, or use fresh ones if state leaks
		// For this structure, mockService is shared, so its expectations should be set per test.
		// mockAuth.UserIDToSet controls what the mock auth middleware does.

		userID := "test-user-123"
		mockAuth.UserIDToSet = userID // Configure mock auth to set this userID

		expectedSummary := &entity_dashboard.DashboardSummary{
			SummaryCards: entity_dashboard.SummaryCards{TotalBalance: 1000},
		}
		mockService.On("GetDashboardSummary", mock.Anything, userID).Return(expectedSummary, nil).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/dashboard/summary", nil)
		// If specific headers are needed by MockHeaderMiddleware or auth, set them here.

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseBody entity_dashboard.DashboardSummary
		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
		assert.NoError(t, err)
		assert.Equal(t, *expectedSummary, responseBody)

		mockService.AssertExpectations(t)
	})

	t.Run("Service Returns Error", func(t *testing.T) {
		userID := "test-user-456"
		mockAuth.UserIDToSet = userID

		serviceError := errors.New("internal service error")
		mockService.On("GetDashboardSummary", mock.Anything, userID).Return(nil, serviceError).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/dashboard/summary", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		// Check error response body if your handler formats it
		var errorResponse map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Contains(t, errorResponse["error"], "Failed to fetch dashboard summary")


		mockService.AssertExpectations(t)
	})

	t.Run("Missing UserID", func(t *testing.T) {
		// Configure mock auth to NOT set a userID
		mockAuth.UserIDToSet = ""
		// No expectation on mockService as it shouldn't be called.

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/dashboard/summary", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		// Check error response body
		var errorResponse map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		// The exact error message depends on the handler's logic for missing userID.
		// Based on the handler code: "User ID is invalid or not provided"
		assert.Contains(t, errorResponse["error"], "User ID is invalid or not provided")

		// Ensure service was not called
		mockService.AssertNotCalled(t, "GetDashboardSummary", mock.Anything, mock.Anything)
	})
}
