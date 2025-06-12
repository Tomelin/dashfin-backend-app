package web_finance

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	financeEntity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	// Service import will be used by the initializer that wires up the actual service
	// "github.com/Tomelin/dashfin-backend-app/internal/core/service/finance"
)

// DashboardServiceInterface defines the methods our handler expects from the dashboard service.
type DashboardServiceInterface interface {
	GetDashboardData(ctx context.Context) (*financeEntity.Dashboard, error)
}

// DashboardHandler handles HTTP requests for dashboard data using Gin.
type DashboardHandler struct {
	service DashboardServiceInterface
	// router  *gin.RouterGroup // Not strictly needed if setupRoutes is called by initializer
}

// NewDashboardHandler creates a new DashboardHandler instance.
// This is kept for direct instantiation, e.g. in tests without full router setup.
func NewDashboardHandler(svc DashboardServiceInterface) *DashboardHandler {
	return &DashboardHandler{
		service: svc,
	}
}

// InitializeDashboardHandler sets up the dashboard handler with its routes.
// This function would be called from the main application setup.
// It mirrors the pattern seen in BankAccountHandler.
func InitializeDashboardHandler(
	svc DashboardServiceInterface,
	routerGroup *gin.RouterGroup,
	// authMiddleware gin.HandlerFunc, // Assuming a common auth middleware
	middleware ...gin.HandlerFunc, // To match BankAccountHandler pattern
) *DashboardHandler {
	handler := &DashboardHandler{
		service: svc,
	}

	handler.setupRoutes(routerGroup, middleware...)
	return handler
}

func (h *DashboardHandler) setupRoutes(
	routerGroup *gin.RouterGroup,
	// authMiddleware gin.HandlerFunc,
	middleware ...gin.HandlerFunc,
) {
	// Apply middleware to the route
	// The middleware slice is already prepared by the caller of setupRoutes
	// or directly in InitializeDashboardHandler.

	// Example: routerGroup.GET("/dashboard", authMiddleware, h.GetDashboard)
	// Using variadic middleware:
	// dashboardRoute := routerGroup.Group("/dashboard") // Create a group for /dashboard
	// {
		// Apply middleware to all routes in this group, or to specific routes.
		// If middleware should apply to the GET route:
		// dashboardRoute.GET("", append(middleware, h.GetDashboard)...)
		// If no specific sub-routes under /dashboard, can do:
		// routerGroup.GET("/dashboard", append(middleware, h.GetDashboard)...)
		// For consistency with BankAccountHandler that sets up routes like /finance/bank-accounts directly on the passed group:
	// }
	// Let's assume the routerGroup passed is already the correct base (e.g., /v1 or /api)
	// and we just add /dashboard to it.
	// If routerGroup is /api/v1, this will make it /api/v1/dashboard
	// If BankAccountHandler adds "/finance/bank-accounts", this should add "/dashboard" or "/finance/dashboard"
	// Let's use "/finance/dashboard" for consistency if it's finance related.

	// Final decision: make it /finance/dashboard to group with other finance routes
	// All middleware passed to InitializeDashboardHandler will be applied.
	routerGroup.GET("/finance/dashboard", append(middleware, h.GetDashboard)...)
}


// GetDashboard is the Gin HTTP handler for GET /dashboard requests.
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	dashboardData, err := h.service.GetDashboardData(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve dashboard data"})
		return
	}
	c.JSON(http.StatusOK, dashboardData)
}
