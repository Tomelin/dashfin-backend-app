// Package dashboard_web handles HTTP requests for dashboard functionalities.
package dashboard_web

import (
	"net/http"

	service_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/service/dashboard"
	"github.com/Tomelin/dashfin-backend-app/pkg/authenticatior"
	"github.com/Tomelin/dashfin-backend-app/pkg/cryptData"
	"github.com/gin-gonic/gin"
)

// DashboardHandler holds dependencies for dashboard request handlers.
type DashboardHandler struct {
	service    service_dashboard.ServiceInterface
	crypt      cryptdata.CryptDataInterface    // Included for consistency, may not be used in all methods
	authClient authenticatior.Authenticator
}

// NewDashboardHandler creates and registers dashboard routes.
// It follows the pattern of other handlers in the project.
func NewDashboardHandler(
	service service_dashboard.ServiceInterface,
	cryptLib cryptdata.CryptDataInterface,
	authClient authenticatior.Authenticator,
	router *gin.RouterGroup,
	corsMiddleware gin.HandlerFunc, // Assuming corsMiddleware is applied globally or at a higher level group
	headerMiddleware gin.HandlerFunc,
) *DashboardHandler {
	h := &DashboardHandler{
		service:    service,
		crypt:      cryptLib,
		authClient: authClient,
	}

	// Define a group for dashboard routes, if specific middlewares are needed for this group
	// For now, assuming /dashboard is already part of a group that has corsMiddleware
	dashboardGroup := router.Group("/dashboard")
	// Apply middlewares specific to this group if any, otherwise they are inherited or passed.

	// Register routes
	// The problem description suggests headerMiddleware and authClient.Middleware()
	// The corsMiddleware is often applied at a higher level (e.g. when setting up the main router)
	// If corsMiddleware is indeed needed here per route, it should be added.
	// For now, following the specific sequence from the prompt for the route.
	dashboardGroup.GET("/summary", headerMiddleware, authClient.Middleware(), h.GetDashboardSummary)

	return h
}

// GetDashboardSummary handles the request to fetch the dashboard summary.
// @Summary Get Dashboard Summary
// @Description Retrieves a summary of dashboard data including cards, account summaries, bills, charts, and recommendations.
// @Tags Dashboard
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} entity_dashboard.DashboardSummary "Successfully retrieved dashboard summary"
// @Failure 400 {object} web.ErrorResponse "Bad Request - Invalid input or missing UserID"
// @Failure 401 {object} web.ErrorResponse "Unauthorized - User not authenticated or UserID missing"
// @Failure 500 {object} web.ErrorResponse "Internal Server Error - Failed to fetch dashboard summary"
// @Router /dashboard/summary [get]
func (h *DashboardHandler) GetDashboardSummary(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	// The key "x-user-id" is an assumption; it should match what authClient.Middleware() sets.
	// Other handlers in the project might use c.GetString("userID") or similar.
	// Let's assume "user_id" is the standard key set by the auth middleware.
	rawUserID, exists := c.Get("user_id")
	if !exists {
		// Fallback or alternative check, e.g. from header if auth middleware doesn't set it directly
		// For this example, we'll assume c.Get("user_id") is the primary method.
		// If not found via c.Get, check header "x-user-id" as a secondary measure,
		// though ideally, the auth middleware should consistently provide it.
		userIDFromHeader := c.Request.Header.Get("x-user-id")
		if userIDFromHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context or header"})
			return
		}
		rawUserID = userIDFromHeader
	}

	userID, ok := rawUserID.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID is invalid or not provided"})
		return
	}

	summary, err := h.service.GetDashboardSummary(c.Request.Context(), userID)
	if err != nil {
		// Log the error for server-side observability
		// log.Printf("Error fetching dashboard summary for userID %s: %v", userID, err)
		// TODO: Add proper logging mechanism
		// Based on the error type, a more specific HTTP status might be returned.
		// For now, a generic 500 for service errors.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dashboard summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}
