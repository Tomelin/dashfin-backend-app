package web_finance

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	financeEntity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	// Assuming DashboardServiceInterface is defined in the same package as DashboardHandler
	// or we define/import it appropriately. The handler itself defines it locally.
)

// MockDashboardService is a mock for the DashboardServiceInterface used by the handler.
type MockDashboardService struct {
	mock.Mock
}

func (m *MockDashboardService) GetDashboardData(ctx context.Context) (*financeEntity.Dashboard, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*financeEntity.Dashboard), args.Error(1)
}

func TestDashboardHandler_GetDashboard_Nominal(t *testing.T) {
	// Setup Gin router and recorder for testing
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Mock request (not strictly necessary for this handler if no params/body read)
	req, _ := http.NewRequest(http.MethodGet, "/dashboard", nil)
	// If UserID needs to be in context for the service (it does, via service's own extraction)
	// Here we are testing the handler; the service is mocked.
	// The handler passes c.Request.Context() to the service.
	// We can set a dummy UserID in the request's context if the mock needs to assert it,
	// but the mock here doesn't specifically check context values, only that it's called.
	// For a more thorough test, the mock could assert ctx.Value("UserID").
	ctxForService := context.WithValue(context.Background(), "UserID", "test-handler-user")
	c.Request = req.WithContext(ctxForService)


	mockService := new(MockDashboardService)
	dashboardHandler := NewDashboardHandler(mockService)

	expectedDashboard := &financeEntity.Dashboard{
		SummaryCards: financeEntity.SummaryCards{TotalBalance: 12345.67},
		// ... other fields as needed for assertion
	}

	// Expect GetDashboardData to be called on the service
	// The context passed to the service mock will be c.Request.Context()
	mockService.On("GetDashboardData", c.Request.Context()).Return(expectedDashboard, nil).Once()

	// Perform the request
	dashboardHandler.GetDashboard(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var responseDashboard financeEntity.Dashboard
	err := json.Unmarshal(w.Body.Bytes(), &responseDashboard)
	assert.NoError(t, err)
	assert.Equal(t, expectedDashboard.SummaryCards.TotalBalance, responseDashboard.SummaryCards.TotalBalance)

	mockService.AssertExpectations(t)
}

func TestDashboardHandler_GetDashboard_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/dashboard", nil)
	ctxForService := context.WithValue(context.Background(), "UserID", "test-handler-user-err")
	c.Request = req.WithContext(ctxForService)


	mockService := new(MockDashboardService)
	dashboardHandler := NewDashboardHandler(mockService)

	serviceError := errors.New("internal service error")
	// Expect GetDashboardData to be called and return an error
	mockService.On("GetDashboardData", c.Request.Context()).Return(nil, serviceError).Once()

	// Perform the request
	dashboardHandler.GetDashboard(c)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var errorResponse map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse["error"], "Failed to retrieve dashboard data")

	mockService.AssertExpectations(t)
}
