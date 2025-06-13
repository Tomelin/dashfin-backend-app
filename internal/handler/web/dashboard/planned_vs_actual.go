package web_dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	entity_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"   // entity_dashboard
	service_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/service/dashboard" // service_dashboard
	"github.com/Tomelin/dashfin-backend-app/internal/handler/web"
	"github.com/Tomelin/dashfin-backend-app/pkg/authenticatior"
	"github.com/Tomelin/dashfin-backend-app/pkg/cryptData"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// PlannedVsActualHandler handles HTTP requests for the planned versus actual expenses feature.
// It depends on a FinancialServiceInterface for business logic, an Authenticator for user authentication,
// and a CryptDataInterface for payload encryption.
type PlannedVsActualHandler struct {
	service     service_dashboard.FinancialServiceInterface // Service to handle business logic.
	authClient  authenticatior.Authenticator                // Client for authenticating users.
	encryptData cryptdata.CryptDataInterface                // Service for encrypting response payloads.
	validate    *validator.Validate                         // Validator for request data.
}

// ErrorResponse defines the standard JSON structure for error responses from this handler.
// Note: This is a local definition. If a global error response structure is available
// (e.g., in a `web` package), that should be preferred for consistency.
type ErrorResponse struct {
	Error string `json:"error"` // Error message.
}

// InitializePlannedVsActualHandler creates a new PlannedVsActualHandler, initializes it with
// necessary dependencies (service, auth client, encryption service), and sets up its routes
// within the provided Gin router group.
//
// Parameters:
//   - svc: The FinancialServiceInterface implementation.
//   - authClient: The Authenticator implementation.
//   - encryptData: The CryptDataInterface for response encryption. Can be nil if encryption is not used.
//   - routerGroup: The Gin router group to which the handler's routes will be added.
//   - middleware: A variadic slice of Gin middleware to be applied to the handler's routes.
//
// Returns:
//   - A pointer to the newly created PlannedVsActualHandler.
func InitializePlannedVsActualHandler(
	svc service_dashboard.FinancialServiceInterface,
	authClient authenticatior.Authenticator,
	encryptData cryptdata.CryptDataInterface, // Service for encrypting response data.
	routerGroup *gin.RouterGroup, // Gin router group for path registration.
	middleware ...gin.HandlerFunc, // Middleware functions to apply to the routes.
) *PlannedVsActualHandler {
	handler := &PlannedVsActualHandler{
		service:     svc,
		authClient:  authClient,
		encryptData: encryptData,
		validate:    validator.New(), // Initialize a new validator instance.
	}
	handler.setupRoutes(routerGroup, middleware...) // Setup routes for this handler.
	return handler
}

// setupRoutes configures the HTTP routes for the PlannedVsActualHandler.
// It defines a GET endpoint at "/dashboard/planned-vs-actual" relative to the provided routerGroup.
// Applied middleware is passed to the route group and specific GET handler.
func (h *PlannedVsActualHandler) setupRoutes(
	routerGroup *gin.RouterGroup, // The parent router group.
	middleware ...gin.HandlerFunc, // Middleware to apply to these routes.
) {
	// Create a new group for /dashboard/planned-vs-actual endpoint
	// e.g. if routerGroup is /api/v1, this becomes /api/v1/dashboard/planned-vs-actual
	plannedVsActualRoutes := routerGroup.Group("/dashboard/planned-vs-actual")

	// Apply all provided middleware to this specific group of routes.
	for _, mw := range middleware {
		plannedVsActualRoutes.Use(mw)
	}

	// Define the GET route.
	// Note: The pattern of applying middleware both to the group and then again here
	// using append(middleware, h.GetPlannedVsActual)... might be redundant if middleware
	// are already applied by plannedVsActualRoutes.Use(mw).
	// Kept for consistency if this is an established pattern in the project.
	// Typically, group-level middleware is sufficient.
	plannedVsActualRoutes.GET("", append(middleware, h.GetPlannedVsActual)...)
}

// GetPlannedVsActual is the Gin handler function for the GET /dashboard/planned-vs-actual endpoint.
// It processes requests for comparing planned versus actual expenses.
//
// Workflow:
// 1. Authenticates the user by validating headers via `web.GetRequiredHeaders`.
// 2. Binds and validates query parameters (Month, Year) from the request.
//    - Performs struct tag validation (e.g., month range 1-12).
//    - Performs custom validation for the year (e.g., must be within a sensible range like 2020 to currentYear+1).
// 3. Calls the underlying financial service (`h.service.GetPlannedVsActual`) to fetch and compute the data.
// 4. Handles service responses:
//    - If the service returns an error, responds with HTTP 500 Internal Server Error.
//    - If the service returns no data (empty slice), responds with HTTP 404 Not Found and an empty JSON array `[]`.
//    - If data is returned successfully:
//        - If an encryption service (`h.encryptData`) is configured, the response payload is encrypted
//          and returned within a `{"payload": "encrypted_string"}` structure.
//        - Otherwise, the data is returned as unencrypted JSON.
// 5. Responds with appropriate HTTP status codes and JSON payloads for errors or success.
func (h *PlannedVsActualHandler) GetPlannedVsActual(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Printf("Error getting required headers for PlannedVsActual: %v", err)
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized: " + err.Error()})
		return
	}

	var req entity_dashboard.PlannedVsActualRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("Error binding query for PlannedVsActual: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request parameters: " + err.Error()})
		return
	}

	// Validate struct tags
	if err := h.validate.Struct(req); err != nil {
		log.Printf("Validation error for PlannedVsActual request (tags): %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid data: " + err.Error()})
		return
	}

	// Custom validation for Year (if provided and not 0, which means default to current year in service)
	if req.Year != 0 {
		currentYear := time.Now().Year()
		if req.Year < 2020 || req.Year > currentYear+1 {
			errMsg := fmt.Sprintf("Year must be between 2020 and %d", currentYear+1)
			log.Printf("Validation error for PlannedVsActual request (year range): %s for input year %d", errMsg, req.Year)
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: errMsg})
			return
		}
	}
	// Month validation (1-12) is handled by struct tags if req.Month is not 0 (omitempty).

	requestCtx := c.Request.Context()
	// Add token to context if any deeper service call might need it (optional, userID is passed directly)
	requestCtx = context.WithValue(requestCtx, web.AuthorizationKey("Authorization"), token)

	results, err := h.service.GetPlannedVsActual(requestCtx, userID, req)
	if err != nil {
		log.Printf("Error from GetPlannedVsActual service for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve planned vs actual data."})
		return
	}

	if len(results) == 0 {
		// Spec: "404: Nenhum dado encontrado (retornar array vazio [])"
		log.Printf("No data found for PlannedVsActual for user %s, query %+v. Returning 404.", userID, req)
		c.JSON(http.StatusNotFound, make([]entity_dashboard.PlannedVsActualCategory, 0))
		return
	}

	if h.encryptData != nil {
		log.Printf("Encrypting response for PlannedVsActual for user %s.", userID)
		responseBytes, err := json.Marshal(results)
		if err != nil {
			log.Printf("Error marshalling results for PlannedVsActual user %s: %v", userID, err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Error preparing response."})
			return
		}
		encryptedResult, err := h.encryptData.EncryptPayload(responseBytes)
		if err != nil {
			log.Printf("Error encrypting response for PlannedVsActual user %s: %v", userID, err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Error securing response."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"payload": encryptedResult})
	} else {
		log.Printf("Returning unencrypted response for PlannedVsActual for user %s.", userID)
		c.JSON(http.StatusOK, results)
	}
}
