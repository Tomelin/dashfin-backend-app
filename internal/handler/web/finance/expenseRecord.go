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

// ExpenseRecordHandlerInterface defines the HTTP handler operations for ExpenseRecords.
type ExpenseRecordHandlerInterface interface {
	CreateExpenseRecord(c *gin.Context)
	GetExpenseRecordByID(c *gin.Context)
	GetExpenseRecords(c *gin.Context)
	GetExpenseRecordsByFilter(c *gin.Context) // Added for filtering
	UpdateExpenseRecord(c *gin.Context)
	DeleteExpenseRecord(c *gin.Context)
}

// ExpenseRecordHandler handles HTTP requests for ExpenseRecords.
type ExpenseRecordHandler struct {
	service     entity_finance.ExpenseRecordServiceInterface
	router      *gin.RouterGroup // Keep if you intend to use it directly, otherwise can be passed in setupRoutes
	encryptData cryptdata.CryptDataInterface
	authClient  authenticatior.Authenticator
}

// InitializeExpenseRecordHandler creates a new ExpenseRecordHandler and sets up routes.
func InitializeExpenseRecordHandler(
	svc entity_finance.ExpenseRecordServiceInterface,
	encryptData cryptdata.CryptDataInterface,
	authClient authenticatior.Authenticator,
	routerGroup *gin.RouterGroup,
	middleware ...gin.HandlerFunc,
) ExpenseRecordHandlerInterface {
	handler := &ExpenseRecordHandler{
		service:     svc,
		encryptData: encryptData,
		authClient:  authClient,
		// router: routerGroup, // Not strictly needed if routerGroup is only used in setupRoutes
	}

	handler.setupRoutes(routerGroup, middleware...)
	return handler
}

func (h *ExpenseRecordHandler) setupRoutes(routerGroup *gin.RouterGroup, middleware ...gin.HandlerFunc) {
	// Apply provided middleware to the group or individual routes
	// For simplicity, applying to all routes here.
	// You might want more granular control.
	// Example: financeRoutes := routerGroup.Group("/finance/expense-records", middleware...)

	financeRoutes := routerGroup.Group("/finance/expenses")
	for _, mw := range middleware {
		financeRoutes.Use(mw)
	}

	financeRoutes.POST("", h.CreateExpenseRecord)
	financeRoutes.GET("/:id", h.GetExpenseRecordByID)
	financeRoutes.GET("", h.GetExpenseRecords)                 // Get all for user
	financeRoutes.POST("/filter", h.GetExpenseRecordsByFilter) // Route for filtered GET
	financeRoutes.PUT("/:id", h.UpdateExpenseRecord)
	financeRoutes.DELETE("/:id", h.DeleteExpenseRecord)
}

// CreateExpenseRecord handles the creation of a new expense record.
func (h *ExpenseRecordHandler) CreateExpenseRecord(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Printf("Error getting required headers: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var payload cryptdata.CryptData
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("Error binding JSON payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	decryptedData, err := h.encryptData.PayloadData(payload.Payload)
	if err != nil {
		log.Printf("Error decrypting payload data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing request data: " + err.Error()})
		return
	}

	var expenseRecord entity_finance.ExpenseRecord
	if err := json.Unmarshal(decryptedData, &expenseRecord); err != nil {
		log.Printf("Error unmarshalling decrypted data to ExpenseRecord: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format: " + err.Error()})
		return
	}
	log.Println("id ==> ", expenseRecord.ID)
	log.Println("category ==> ", expenseRecord.Category)
	log.Println("subcategory ==> ", expenseRecord.Subcategory)
	log.Println("dueDate ==> ", expenseRecord.DueDate)
	log.Println("paymentDate ==> ", *expenseRecord.PaymentDate)
	log.Println("amount ==> ", expenseRecord.Amount)
	log.Println("bankPaidFrom ==> ", expenseRecord.BankPaidFrom)
	log.Println("customBankName ==> ", expenseRecord.CustomBankName)
	log.Println("description ==> ", expenseRecord.Description)
	log.Println("isRecurring ==> ", expenseRecord.IsRecurring)
	log.Println("recurrenceCount ==> ", expenseRecord.RecurrenceCount)
	log.Println("createdAt ==> ", expenseRecord.CreatedAt)
	log.Println("updatedAt ==> ", expenseRecord.UpdatedAt)
	log.Println("userID ==> ", expenseRecord.UserID)

	// Set UserID from authenticated user
	expenseRecord.UserID = userID

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	result, err := h.service.CreateExpenseRecord(ctx, &expenseRecord)
	if err != nil {
		log.Printf("Error creating expense record via service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create expense record: " + err.Error()})
		return
	}

	responseBytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("Error marshalling result to JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error preparing response: " + err.Error()})
		return
	}

	encryptedResult, err := h.encryptData.EncryptPayload(responseBytes)
	if err != nil {
		log.Printf("Error encrypting response payload: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error securing response: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"payload": encryptedResult})
}

// GetExpenseRecordByID handles fetching a single expense record by its ID.
func (h *ExpenseRecordHandler) GetExpenseRecordByID(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Printf("Error getting required headers: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID parameter is required"})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	result, err := h.service.GetExpenseRecordByID(ctx, id)
	if err != nil {
		log.Printf("Error getting expense record by ID via service (ID: %s): %v", id, err)
		// Distinguish between "not found" and other errors if possible
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve expense record: " + err.Error()})
		}
		return
	}

	responseBytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("Error marshalling result to JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error preparing response: " + err.Error()})
		return
	}

	encryptedResult, err := h.encryptData.EncryptPayload(responseBytes)
	if err != nil {
		log.Printf("Error encrypting response payload: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error securing response: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": encryptedResult})
}

// GetExpenseRecords handles fetching all expense records for the authenticated usexpenseRecord.
func (h *ExpenseRecordHandler) GetExpenseRecords(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Printf("Error getting required headers: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	results, err := h.service.GetExpenseRecords(ctx)
	if err != nil {
		log.Printf("Error getting expense records via service: %v", err)
		// If service returns "not found" for an empty list, handle it gracefully
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusOK, gin.H{"payload": "[]"}) // Return empty JSON array
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve expense records: " + err.Error()})
		return
	}

	// Handle case where results might be nil from service if no records found (and no error)
	if results == nil {
		results = []entity_finance.ExpenseRecord{} // Ensure a non-nil slice for marshalling
	}

	responseBytes, err := json.Marshal(results)
	if err != nil {
		log.Printf("Error marshalling results to JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error preparing response: " + err.Error()})
		return
	}

	encryptedResult, err := h.encryptData.EncryptPayload(responseBytes)
	if err != nil {
		log.Printf("Error encrypting response payload: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error securing response: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": encryptedResult})
}

// GetExpenseRecordsByFilter handles fetching expense records based on a filtexpenseRecord.
func (h *ExpenseRecordHandler) GetExpenseRecordsByFilter(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Printf("Error getting required headers: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var payload cryptdata.CryptData
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("Error binding JSON payload for filter: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload for filter: " + err.Error()})
		return
	}

	decryptedData, err := h.encryptData.PayloadData(payload.Payload)
	if err != nil {
		log.Printf("Error decrypting payload data for filter: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing request data for filter: " + err.Error()})
		return
	}

	var filter map[string]interface{}
	if err := json.Unmarshal(decryptedData, &filter); err != nil {
		log.Printf("Error unmarshalling decrypted data to filter map: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter data format: " + err.Error()})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	results, err := h.service.GetExpenseRecordsByFilter(ctx, filter)
	if err != nil {
		log.Printf("Error getting expense records by filter via service: %v", err)
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusOK, gin.H{"payload": "[]"}) // Return empty JSON array
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve expense records by filter: " + err.Error()})
		return
	}

	if results == nil {
		results = []entity_finance.ExpenseRecord{}
	}

	responseBytes, err := json.Marshal(results)
	if err != nil {
		log.Printf("Error marshalling filtered results to JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error preparing response: " + err.Error()})
		return
	}

	encryptedResult, err := h.encryptData.EncryptPayload(responseBytes)
	if err != nil {
		log.Printf("Error encrypting response payload: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error securing response: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": encryptedResult})
}

// UpdateExpenseRecord handles updating an existing expense record.
func (h *ExpenseRecordHandler) UpdateExpenseRecord(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Printf("Error getting required headers: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID parameter is required for update"})
		return
	}

	var payload cryptdata.CryptData
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("Error binding JSON payload for update: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload for update: " + err.Error()})
		return
	}

	decryptedData, err := h.encryptData.PayloadData(payload.Payload)
	if err != nil {
		log.Printf("Error decrypting payload data for update: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing request data for update: " + err.Error()})
		return
	}

	var expenseRecord entity_finance.ExpenseRecord
	if err := json.Unmarshal(decryptedData, &expenseRecord); err != nil {
		log.Printf("Error unmarshalling decrypted data to ExpenseRecord for update: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format for update: " + err.Error()})
		return
	}

	// Set UserID from authenticated user for the update payload
	expenseRecord.UserID = userID

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	// The service's UpdateExpenseRecord should handle verifying ownership and setting UpdatedAt.
	// Pass the ID from the path and the unmarshalled data.
	result, err := h.service.UpdateExpenseRecord(ctx, id, &expenseRecord)
	if err != nil {
		log.Printf("Error updating expense record via service (ID: %s): %v", id, err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update expense record: " + err.Error()})
		}
		return
	}

	responseBytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("Error marshalling updated result to JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error preparing response: " + err.Error()})
		return
	}

	encryptedResult, err := h.encryptData.EncryptPayload(responseBytes)
	if err != nil {
		log.Printf("Error encrypting response payload: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error securing response: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": encryptedResult})
}

// DeleteExpenseRecord handles deleting an expense record.
func (h *ExpenseRecordHandler) DeleteExpenseRecord(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Printf("Error getting required headers: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID parameter is required for delete"})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	err = h.service.DeleteExpenseRecord(ctx, id)
	if err != nil {
		log.Printf("Error deleting expense record via service (ID: %s): %v", id, err)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete expense record: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
