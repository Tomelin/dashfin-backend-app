package web_finance // Changed package name

import (
	// "errors" // For placeholder function - removed as not used by current placeholder
	"net/http"

	"github.com/gin-gonic/gin"
	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	service_finance "github.com/Tomelin/dashfin-backend-app/internal/core/service/finance"
)

// SpendingPlanHandler handles HTTP requests for spending plans.
type SpendingPlanHandler struct {
	service service_finance.SpendingPlanService
}

// NewSpendingPlanHandler creates a new SpendingPlanHandler.
func NewSpendingPlanHandler(service service_finance.SpendingPlanService) *SpendingPlanHandler {
	return &SpendingPlanHandler{service: service}
}

// RegisterSpendingPlanRoutes sets up the routes for spending plan operations under the given router group.
func (h *SpendingPlanHandler) RegisterSpendingPlanRoutes(group *gin.RouterGroup) {
	spendingPlanGroup := group.Group("/spending-plan") // Changed to make routes /spending-plan
	{
		spendingPlanGroup.GET("", h.GetSpendingPlan)    // GET /api/finance/spending-plan
		spendingPlanGroup.PUT("", h.SaveSpendingPlan)   // PUT /api/finance/spending-plan
	}
}


// getUserIDFromContext is a placeholder function to extract user ID.
// In a real application, this would come from a JWT token or session validated by middleware.
func getUserIDFromContext(c *gin.Context) (string, error) {
	// Example: Retrieve from a context value set by auth middleware
	// userID, exists := c.Get("userID")
	// if !exists {
	// 	return "", errors.New("userID not found in context")
	// }
	// return userID.(string), nil

	// For now, using a hardcoded value for development/testing:
	// It's better to get it from a header for temporary testing if possible
	// testUserID := c.GetHeader("X-Test-User-ID")
	// if testUserID != "" {
	// 	return testUserID, nil
	// }
	return "temp-user-id-from-handler", nil // Placeholder
}

// GetSpendingPlan handles the GET /spending-plan request.
func (h *SpendingPlanHandler) GetSpendingPlan(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "details": err.Error()})
		return
	}

	// Use c.Request.Context() for the service call
	plan, err := h.service.GetSpendingPlan(c.Request.Context(), userID)
	if err != nil {
		// Differentiate between "not found" (which might not be an error from service, but a nil plan)
		// and actual server errors.
		// For now, assuming service might return a specific error for not found, or just (nil, nil).
		// Based on current service impl, (nil, error) or (nil,nil) if not found and repo returns (nil,nil)
		// Let's assume a generic error from service indicates a problem.
		// A more robust error handling would involve custom error types.
		// If errors.Is(err, some_specific_not_found_error_type) {
		//  c.JSON(http.StatusNotFound, gin.H{"error": "Spending plan not found"})
		//  return
		// }
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve spending plan", "details": err.Error()})
		return
	}

	if plan == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Spending plan not found for this user"})
		return
	}

	c.JSON(http.StatusOK, plan)
}

// SaveSpendingPlan handles the PUT /spending-plan request.
func (h *SpendingPlanHandler) SaveSpendingPlan(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "details": err.Error()})
		return
	}

	var requestBody entity_finance.SpendingPlan
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// The service's SaveSpendingPlan will set UserID, CreatedAt, UpdatedAt.
	// requestBody here primarily brings MonthlyIncome and CategoryBudgets from the user input.
	// UserID from context takes precedence.
	// The service method SaveSpendingPlan(ctx, userID, planData) will handle existingPlan.UserID = userID
	// For a new plan, it also sets UserID.
	// So, requestBody.UserID from JSON is effectively ignored by the service if userID from context is used.

	savedPlan, err := h.service.SaveSpendingPlan(c.Request.Context(), userID, &requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save spending plan", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, savedPlan)
}
