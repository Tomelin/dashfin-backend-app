package web_finance // Changed package name

import (
	// "errors" // For placeholder function - removed as not used by current placeholder
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	web "github.com/Tomelin/dashfin-backend-app/internal/handler/web"
	"github.com/Tomelin/dashfin-backend-app/pkg/authenticatior"
	cryptdata "github.com/Tomelin/dashfin-backend-app/pkg/cryptData"
	"github.com/gin-gonic/gin"
)

type SpendingPlanHandlerInterface interface {
	GetSpendingPlan(c *gin.Context)
	SaveSpendingPlan(c *gin.Context)
}

// SpendingPlanHandler handles HTTP requests for spending plans.
type SpendingPlanHandler struct {
	service     entity_finance.SpendingPlanServiceInterface
	encryptData cryptdata.CryptDataInterface
	authClient  authenticatior.Authenticator
}

// InitializeSpendingPlanHandler creates a new SpendingPlanHandler.
func InitializeSpendingPlanHandler(
	service entity_finance.SpendingPlanServiceInterface,
	encryptData cryptdata.CryptDataInterface,
	authClient authenticatior.Authenticator,
	routerGroup *gin.RouterGroup,
	middleware ...gin.HandlerFunc,
) *SpendingPlanHandler {

	handler := &SpendingPlanHandler{
		service:     service,
		encryptData: encryptData,
		authClient:  authClient,
	}

	handler.setupRoutes(routerGroup, middleware...)
	return handler
}

// RegisterSpendingPlanRoutes sets up the routes for spending plan operations under the given router group.
func (h *SpendingPlanHandler) setupRoutes(routerGroup *gin.RouterGroup, middleware ...gin.HandlerFunc) {

	spendingPlanGroup := routerGroup.Group("/finance/spending-plan")
	for _, mw := range middleware {
		spendingPlanGroup.Use(mw)
	}

	spendingPlanGroup.GET("", h.GetSpendingPlan)
	spendingPlanGroup.PUT("", h.SaveSpendingPlan)
}

// GetSpendingPlan handles the GET /spending-plan request.
func (h *SpendingPlanHandler) GetSpendingPlan(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Printf("Error getting required headers: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	result, err := h.service.GetSpendingPlan(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve income record: " + err.Error()})
		}
		return
	}

	responseBytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("Error marshalling result to JSON for GetIncomeRecordByID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error preparing response: " + err.Error()})
		return
	}

	encryptedResult, err := h.encryptData.EncryptPayload(responseBytes)
	if err != nil {
		log.Printf("Error encrypting response payload for GetIncomeRecordByID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error securing response: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": encryptedResult})
}

// SaveSpendingPlan handles the PUT /spending-plan request.
func (h *SpendingPlanHandler) SaveSpendingPlan(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Printf("Error getting required headers: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var payload cryptdata.CryptData
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("Error binding JSON payload for CreateIncomeRecord: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	decryptedData, err := h.encryptData.PayloadData(payload.Payload)
	if err != nil {
		log.Printf("Error decrypting payload data for CreateIncomeRecord: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing request data: " + err.Error()})
		return
	}

	var spendingRecord entity_finance.SpendingPlan
	if err := json.Unmarshal(decryptedData, &spendingRecord); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format: " + err.Error()})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	spendingRecord.UserID = userID

	savedPlan, err := h.service.UpdateSpendingPlan(ctx, &spendingRecord)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save spending plan", "details": err.Error()})
		return
	}

	responseBytes, err := json.Marshal(savedPlan)
	if err != nil {
		log.Printf("Error marshalling result to JSON for CreateIncomeRecord: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error preparing response: " + err.Error()})
		return
	}

	encryptedResult, err := h.encryptData.EncryptPayload(responseBytes)
	if err != nil {
		log.Printf("Error encrypting response payload for CreateIncomeRecord: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error securing response: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"payload": encryptedResult})
}
