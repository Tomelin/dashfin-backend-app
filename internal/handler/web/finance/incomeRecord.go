package web_finance

import (
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

// IncomeRecordHandlerInterface defines the HTTP handler operations for IncomeRecords.
type IncomeRecordHandlerInterface interface {
	CreateIncomeRecord(c *gin.Context)
	GetIncomeRecordByID(c *gin.Context)
	GetIncomeRecords(c *gin.Context)
	UpdateIncomeRecord(c *gin.Context)
	DeleteIncomeRecord(c *gin.Context)
}

// IncomeRecordHandler handles HTTP requests for IncomeRecords.
type IncomeRecordHandler struct {
	service     entity_finance.IncomeRecordServiceInterface
	encryptData cryptdata.CryptDataInterface
	authClient  authenticatior.Authenticator
}

// InitializeIncomeRecordHandler creates a new IncomeRecordHandler and sets up routes.
func InitializeIncomeRecordHandler(
	svc entity_finance.IncomeRecordServiceInterface,
	encryptData cryptdata.CryptDataInterface,
	authClient authenticatior.Authenticator,
	routerGroup *gin.RouterGroup,
	middleware ...gin.HandlerFunc,
) IncomeRecordHandlerInterface {
	handler := &IncomeRecordHandler{
		service:     svc,
		encryptData: encryptData,
		authClient:  authClient,
	}

	handler.setupRoutes(routerGroup, middleware...)
	return handler
}

func (h *IncomeRecordHandler) setupRoutes(routerGroup *gin.RouterGroup, middleware ...gin.HandlerFunc) {
	// Path from YAML: /api/finance/income
	// The routerGroup is likely already /api, so we add /finance/income
	incomeRoutes := routerGroup.Group("/finance/income")
	for _, mw := range middleware {
		incomeRoutes.Use(mw)
	}

	incomeRoutes.POST("", h.CreateIncomeRecord)
	incomeRoutes.GET("", h.GetIncomeRecords)
	incomeRoutes.GET("/:incomeId", h.GetIncomeRecordByID)
	incomeRoutes.PUT("/:incomeId", h.UpdateIncomeRecord)
	incomeRoutes.DELETE("/:incomeId", h.DeleteIncomeRecord)
}

// CreateIncomeRecord handles the creation of a new income record.
func (h *IncomeRecordHandler) CreateIncomeRecord(c *gin.Context) {
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

	var incomeRecord entity_finance.IncomeRecord
	if err := json.Unmarshal(decryptedData, &incomeRecord); err != nil {
		log.Printf("Error unmarshalling decrypted data to IncomeRecord: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format: " + err.Error()})
		return
	}

	// UserID will be set by the service from context.
	// incomeRecord.UserID = userID // Service layer handles setting UserID from context

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	result, err := h.service.CreateIncomeRecord(ctx, &incomeRecord)
	if err != nil {
		log.Printf("Error creating income record via service: %v", err)
		// Consider more specific error codes based on err type if possible
		if strings.Contains(err.Error(), "validation failed") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create income record: " + err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create income record: " + err.Error()})
		}
		return
	}

	responseBytes, err := json.Marshal(result)
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

// GetIncomeRecordByID handles fetching a single income record by its ID.
func (h *IncomeRecordHandler) GetIncomeRecordByID(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Printf("Error getting required headers: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("incomeId") // Match path parameter from setupRoutes
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Income ID parameter is required"})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	result, err := h.service.GetIncomeRecordByID(ctx, id)
	if err != nil {
		log.Printf("Error getting income record by ID via service (ID: %s): %v", id, err)
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

// GetIncomeRecords handles fetching income records with optional filters.
func (h *IncomeRecordHandler) GetIncomeRecords(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Printf("Error getting required headers: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract query parameters as defined in income-records.yaml
	description := c.Query("description")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	sortKey := c.Query("sortKey")
	sortDirection := c.Query("sortDirection")

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	// The service's GetIncomeRecords expects userID as a parameter.
	// We use the authenticated userID from headers.

	record := entity_finance.GetIncomeRecordsQueryParameters{
		UserID:        userID,
		Description:   &description,
		StartDate:     &startDate,
		EndDate:       &endDate,
		SortKey:       &sortKey,
		SortDirection: &sortDirection,
	}

	results, err := h.service.GetIncomeRecords(ctx, &record)
	if err != nil {
		log.Printf("Error getting income records via service: %v", err)
		if strings.Contains(err.Error(), "not found") { // Or other specific non-fatal errors
			// Based on expenseRecord, "not found" for a list should be an empty list, not error.
			// However, the service/repo might not distinguish "no records found" from other errors.
			// For now, let's assume an empty list is handled by `results == nil` check below.
			// If service explicitly returns "not found" for empty list matching criteria, this could be StatusOK with empty encrypted payload
		}
		if strings.Contains(err.Error(), "invalid query parameters") || strings.Contains(err.Error(), "invalid sortDirection value") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve income records: " + err.Error()})
			return
		} else if err != nil { // Catch-all for other errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve income records: " + err.Error()})
			return
		}
	}

	if results == nil { // Ensure we always return a list, even if empty
		results = []entity_finance.IncomeRecord{}
	}

	responseBytes, err := json.Marshal(results)
	if err != nil {
		log.Printf("Error marshalling results to JSON for GetIncomeRecords: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error preparing response: " + err.Error()})
		return
	}

	encryptedResult, err := h.encryptData.EncryptPayload(responseBytes)
	if err != nil {
		log.Printf("Error encrypting response payload for GetIncomeRecords: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error securing response: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": encryptedResult})
}

// UpdateIncomeRecord handles updating an existing income record.
func (h *IncomeRecordHandler) UpdateIncomeRecord(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Printf("Error getting required headers: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("incomeId") // Match path parameter
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Income ID parameter is required for update"})
		return
	}

	var payload cryptdata.CryptData
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("Error binding JSON payload for UpdateIncomeRecord: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload for update: " + err.Error()})
		return
	}

	decryptedData, err := h.encryptData.PayloadData(payload.Payload)
	if err != nil {
		log.Printf("Error decrypting payload data for UpdateIncomeRecord: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing request data for update: " + err.Error()})
		return
	}

	var incomeRecord entity_finance.IncomeRecord
	if err := json.Unmarshal(decryptedData, &incomeRecord); err != nil {
		log.Printf("Error unmarshalling decrypted data to IncomeRecord for update: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format for update: " + err.Error()})
		return
	}

	// UserID will be set by the service from context.
	// incomeRecord.UserID = userID // Service handles setting UserID from context

	ctx := context.WithValue(c.Request.Context(), "", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	result, err := h.service.UpdateIncomeRecord(ctx, id, &incomeRecord)
	if err != nil {
		log.Printf("Error updating income record via service (ID: %s): %v", id, err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "validation failed") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update income record: " + err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update income record: " + err.Error()})
		}
		return
	}

	responseBytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("Error marshalling updated result to JSON for UpdateIncomeRecord: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error preparing response: " + err.Error()})
		return
	}

	encryptedResult, err := h.encryptData.EncryptPayload(responseBytes)
	if err != nil {
		log.Printf("Error encrypting response payload for UpdateIncomeRecord: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error securing response: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": encryptedResult})
}

// DeleteIncomeRecord handles deleting an income record.
func (h *IncomeRecordHandler) DeleteIncomeRecord(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Printf("Error getting required headers: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("incomeId") // Match path parameter
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Income ID parameter is required for delete"})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	err = h.service.DeleteIncomeRecord(ctx, id)
	if err != nil {
		log.Printf("Error deleting income record via service (ID: %s): %v", id, err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete income record: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
