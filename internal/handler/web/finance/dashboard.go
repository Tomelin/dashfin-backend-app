package web_finance

import (
	"context" // Keep for DashboardServiceInterface
	"net/http"

	// Gin framework
	"github.com/gin-gonic/gin"

	// Service and entity imports
	financeEntity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	// "github.com/Tomelin/dashfin-backend-app/internal/core/service/finance" // Service is used via interface
)

// DashboardServiceInterface defines the methods our handler expects from the dashboard service.
type DashboardServiceInterface interface {
	GetDashboardData(ctx context.Context) (*financeEntity.Dashboard, error)
}

// DashboardHandler handles HTTP requests for dashboard data using Gin.
type DashboardHandler struct {
	service DashboardServiceInterface // Use the interface for decoupling
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(svc DashboardServiceInterface) *DashboardHandler {
	return &DashboardHandler{
		service: svc,
	}
}

// GetDashboard is the Gin HTTP handler for GET /dashboard requests.
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	// UserID extraction from context is handled by the DashboardService.
	// Assumes a middleware has populated "UserID" in the gin.Context or its underlying context.
	// The service will use c.Request.Context().Value("UserID").

	dashboardData, err := h.service.GetDashboardData(c.Request.Context()) // Pass the context from Gin's request
	if err != nil {
		// Log the error appropriately using Gin's error handling or standard log
		// c.Error(err) // Example of Gin's error logging
		// log.Printf("Error getting dashboard data: %v", err)

		// Decide on the error message and status code based on the error type
		// For now, a generic server error. More specific error handling can be added.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve dashboard data"})
		return
	}

	c.JSON(http.StatusOK, dashboardData)
}

// Note: Registration of the route (e.g., routerGroup.GET("/dashboard", dashboardHandler.GetDashboard))
// will be done in a different file where the main Gin router is configured, similar to BankAccountHandler.
