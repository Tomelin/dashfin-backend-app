package web_finance_expense

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	web "github.com/Tomelin/dashfin-backend-app/internal/handler/web"
	"github.com/Tomelin/dashfin-backend-app/internal/handler/web/finance/expense/dto"
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
	CreateExpenseByNfceUrl(c *gin.Context)
	ProcessExpenseByNfceUrl(c *gin.Context)
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
	financeRoutes.POST("/process-nfce-url", h.CreateExpenseByNfceUrl)
	financeRoutes.GET("/process", h.ProcessExpenseByNfceUrl)
}

func (h *ExpenseRecordHandler) ProcessExpenseByNfceUrl(c *gin.Context) {
	ctx := context.WithValue(c.Request.Context(), "Authorization", "token")
	ctx = context.WithValue(ctx, "UserID", "userID")

	expenseNfceUrl := entity_finance.ExpenseByNfceUrl{
		NfceUrl:    "https://dfe-portal.svrs.rs.gov.br/Dfe/QrCodeNFce?p=43250400776574163454653020000395661694784220%7C2%7C1%7C1%7Cb5bd8ab6f361bea7d94707cdcacfd96b44b4d42b",
		UserID:     "userID",
		ImportMode: entity_finance.NfceUrlItems,
	}

	_, err := h.service.CreateExpenseByNfceUrl(ctx, &expenseNfceUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create expense record: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"payload": "ok"})
}

func (h *ExpenseRecordHandler) CreateExpenseByNfceUrl(c *gin.Context) {
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

	var expenseNfceUrl entity_finance.ExpenseByNfceUrl
	if err := json.Unmarshal(decryptedData, &expenseNfceUrl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format: " + err.Error()})
		return
	}

	expenseNfceUrl.UserID = userID

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	result, err := h.service.CreateExpenseByNfceUrl(ctx, &expenseNfceUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create expense record: " + err.Error()})
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

// CreateExpenseRecord handles the creation of a new expense record.
func (h *ExpenseRecordHandler) CreateExpenseRecord(c *gin.Context) {
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

	log.Println("Decrypting Data...")
	var expenseData dto.ExpenseRecordDTO
	if err := json.Unmarshal(decryptedData, &expenseData); err != nil {
		log.Println("Error unmarshalling data:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format: " + err.Error()})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	// Set UserID from authenticated user
	expenseData.UserID = userID

	expenseRecord, err := expenseData.ToEntity()
	if err != nil {
		log.Println("Error converting DTO to entity:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense record data: " + err.Error()})
		return
	}

	log.Println("Creating Expense Record:", expenseRecord)
	result, err := h.service.CreateExpenseRecord(ctx, expenseRecord)
	if err != nil {
		log.Println("Error creating expense record:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create expense record: " + err.Error()})
		return
	}

	expenseResponse := dto.ExpenseRecordDTO{}
	expenseResponse.FromEntity(result)

	responseBytes, err := json.Marshal(expenseResponse)
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

// GetExpenseRecordByID handles fetching a single expense record by its ID.
func (h *ExpenseRecordHandler) GetExpenseRecordByID(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
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
		// Distinguish between "not found" and other errors if possible
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			// c.JSON(http.StatusNoContent, gin.H{"payload": err.Error()})
			// c.JSON(http.StatusNoContent, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve expense record: " + err.Error()})
			return
		}
	}

	expenseResponse := dto.ExpenseRecordDTO{}
	expenseResponse.FromEntity(result)

	responseBytes, err := json.Marshal(expenseResponse)
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

// GetExpenseRecords handles fetching all expense records for the authenticated usexpenseRecord.
func (h *ExpenseRecordHandler) GetExpenseRecords(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	// define variables for filtering
	results := make([]entity_finance.ExpenseRecord, 0)

	if startDate != "" || endDate != "" {
		filter := entity_finance.ExpenseRecordQueryByDate{}
		if startDate != "" {
			filter.StartDate = startDate
		}
		if endDate != "" {
			filter.EndDate = endDate
		}
		// Call the service method that handles filtering
		results, _ = h.service.GetExpenseRecordsByDate(ctx, &filter)

		if results == nil {
			results = []entity_finance.ExpenseRecord{}
		}

	} else {
		// Existing code for getting all records if no filter
		results, _ = h.service.GetExpenseRecords(ctx)
	}

	expenseResponse := make([]dto.ExpenseRecordDTO, 0, len(results))
	for _, record := range results {
		expenseDTO := dto.ExpenseRecordDTO{}
		expenseDTO.FromEntity(&record)
		expenseResponse = append(expenseResponse, expenseDTO)
	}

	responseBytes, err := json.Marshal(expenseResponse)
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

// GetExpenseRecordsByFilter handles fetching expense records based on a filtexpenseRecord.
func (h *ExpenseRecordHandler) GetExpenseRecordsByFilter(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var payload cryptdata.CryptData
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload for filter: " + err.Error()})
		return
	}

	decryptedData, err := h.encryptData.PayloadData(payload.Payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing request data for filter: " + err.Error()})
		return
	}

	var filter map[string]interface{}
	if err := json.Unmarshal(decryptedData, &filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter data format: " + err.Error()})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	results, err := h.service.GetExpenseRecordsByFilter(ctx, filter)
	if err != nil {
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

	expenseResponse := make([]dto.ExpenseRecordDTO, 0, len(results))
	for _, record := range results {
		expenseDTO := dto.ExpenseRecordDTO{}
		expenseDTO.FromEntity(&record)
		expenseResponse = append(expenseResponse, expenseDTO)
	}

	responseBytes, err := json.Marshal(expenseResponse)
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

// UpdateExpenseRecord handles updating an existing expense record.
func (h *ExpenseRecordHandler) UpdateExpenseRecord(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload for update: " + err.Error()})
		return
	}

	decryptedData, err := h.encryptData.PayloadData(payload.Payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing request data for update: " + err.Error()})
		return
	}

	var expenseDTO dto.ExpenseRecordDTO
	if err := json.Unmarshal(decryptedData, &expenseDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format for update: " + err.Error()})
		return
	}

	// Set UserID from authenticated user for the update payload
	expenseDTO.UserID = userID

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	expenseRecord, err := expenseDTO.ToEntity()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense record data for update: " + err.Error()})
		return
	}

	// The service's UpdateExpenseRecord should handle verifying ownership and setting UpdatedAt.
	// Pass the ID from the path and the unmarshalled data.
	result, err := h.service.UpdateExpenseRecord(ctx, id, expenseRecord)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update expense record: " + err.Error()})
			return
		}
	}

	expenseResponse := dto.ExpenseRecordDTO{}
	expenseResponse.FromEntity(result)

	responseBytes, err := json.Marshal(expenseResponse)
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

// DeleteExpenseRecord handles deleting an expense record.
func (h *ExpenseRecordHandler) DeleteExpenseRecord(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
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
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete expense record: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusNoContent, nil)
}
