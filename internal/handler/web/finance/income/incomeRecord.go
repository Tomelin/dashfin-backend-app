package web_finance_income

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	web "github.com/Tomelin/dashfin-backend-app/internal/handler/web"
	"github.com/Tomelin/dashfin-backend-app/internal/handler/web/finance/income/dto"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var payload cryptdata.CryptData
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	decryptedData, err := h.encryptData.PayloadData(payload.Payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing request data: " + err.Error()})
		return
	}

	var incomeRecord dto.IncomeRecordDTO
	if err := json.Unmarshal(decryptedData, &incomeRecord); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format: " + err.Error()})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	incomeRecord.UserID = userID

	income, err := incomeRecord.ToEntity()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid income record data: " + err.Error()})
		return
	}

	income.UserID = userID // Ensure the user ID is set for the new record
	result, err := h.service.CreateIncomeRecord(ctx, income)
	if err != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error preparing response: " + err.Error()})
		return
	}

	encryptedResult, err := h.encryptData.EncryptPayload(responseBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error securing response: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"payload": encryptedResult})
}

// GetIncomeRecordByID handles fetching a single income record by its ID.
func (h *IncomeRecordHandler) GetIncomeRecordByID(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
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
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve income record: " + err.Error()})
		}
		return
	}

	responseBytes, err := json.Marshal(result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error preparing response: " + err.Error()})
		return
	}

	encryptedResult, err := h.encryptData.EncryptPayload(responseBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error securing response: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": encryptedResult})
}

// GetIncomeRecords handles fetching income records with optional filters.
func (h *IncomeRecordHandler) GetIncomeRecords(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
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
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve income records: " + err.Error()})
		}
		return
	}

	incomeRecords := make([]dto.IncomeRecordDTO, 0, len(results))

	for _, income := range results {
		var dtoIncome dto.IncomeRecordDTO
		dtoIncome.FromEntity(&income)
		incomeRecords = append(incomeRecords, dtoIncome)
	}

	responseBytes, err := json.Marshal(incomeRecords)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error preparing response: " + err.Error()})
		return
	}

	encryptedResult, err := h.encryptData.EncryptPayload(responseBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error securing response: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": encryptedResult})
}

// UpdateIncomeRecord handles updating an existing income record.
func (h *IncomeRecordHandler) UpdateIncomeRecord(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload for update: " + err.Error()})
		return
	}

	decryptedData, err := h.encryptData.PayloadData(payload.Payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing request data for update: " + err.Error()})
		return
	}

	var incomeRecord dto.IncomeRecordDTO
	if err := json.Unmarshal(decryptedData, &incomeRecord); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format for update: " + err.Error()})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	incomeRecord.UserID = userID

	income, err := incomeRecord.ToEntity()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid income record data for update: " + err.Error()})
		return
	}

	income.UserID = id // Ensure the ID is set for the update operation

	result, err := h.service.UpdateIncomeRecord(ctx, id, income)
	if err != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error preparing response: " + err.Error()})
		return
	}

	encryptedResult, err := h.encryptData.EncryptPayload(responseBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error securing response: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": encryptedResult})
}

// DeleteIncomeRecord handles deleting an income record.
func (h *IncomeRecordHandler) DeleteIncomeRecord(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
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
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete income record: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
