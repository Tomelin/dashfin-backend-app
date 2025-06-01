package support

import (
	"net/http"
	// "log" // Will be used when calling the service

	"github.com/gin-gonic/gin"
	// "github.com/user/supportservice/internal/platform/web" // For AuthMiddleware if setting up router here
)

// HTTPHandler holds dependencies for the support HTTP handlers.
// For now, it's empty, but it would typically hold a reference to the support service.
// type HTTPHandler struct {
//     SupportService *Service // Example
// }

// NewHTTPHandler creates a new HTTPHandler.
// func NewHTTPHandler(service *Service) *HTTPHandler {
//     return &HTTPHandler{SupportService: service}
// }

// CreateSupportRequest handles the creation of a new support request.
// It expects to be used with the AuthMiddleware.
func CreateSupportRequest(c *gin.Context) {
	var req SupportRequest

	// Bind JSON request to the SupportRequest struct and validate
	// based on struct tags (e.g., binding:"required,min=10,max=2000").
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	// Validate Category enum
	if !SupportRequestCategory(req.Category).IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category value", "details": req.Category})
		return
	}

	// Retrieve validated data from context (set by AuthMiddleware)
	firebaseUID, _ := c.Get("firebase_uid")
	appName, _ := c.Get("app_name")
	userIDHeader := c.GetHeader("X-USERID") // Already validated by middleware to match firebaseUID

	// Placeholder for calling the service layer:
	// log.Printf("Received support request from UID: %s (X-USERID: %s) for app: %s", firebaseUID, userIDHeader, appName)
	// log.Printf("Request details: Category: %s, Description: %s", req.Category, req.Description)
	// err := h.SupportService.Create(c.Request.Context(), req, firebaseUID.(string), appName.(string))
	// if err != nil {
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create support request", "details": err.Error()})
	//     return
	// }

	c.JSON(http.StatusCreated, gin.H{
		"message": "Support request received successfully.",
		"data": req,
		"user_id": userIDHeader, // same as firebaseUID
		"app_name": appName,
	})
}

/*
// Example of how this might be registered in main.go or a routes file:
func RegisterRoutes(router *gin.Engine, authClient *firebase.AuthClient, service *Service) {
	handler := NewHTTPHandler(service)

	// Group for authenticated routes
	// authRequired := router.Group("/") // Or some base path like "/api/v1"
	// authRequired.Use(web.AuthMiddleware(authClient)) // Apply middleware
	// {
	//     authRequired.POST("/support-request", handler.CreateSupportRequest)
	// }
}
*/
