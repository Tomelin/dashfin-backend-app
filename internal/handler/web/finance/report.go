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

type ReportHandlerInterface interface {
	GetSpendingPlan(c *gin.Context)
	SaveSpendingPlan(c *gin.Context)
}

// ReportHandler handles HTTP requests for spending plans.
type ReportHandler struct {
	service     entity_finance.FinancialReportDataServiceInterface
	encryptData cryptdata.CryptDataInterface
	authClient  authenticatior.Authenticator
}

// InitializeReportHandler creates a new ReportHandler.
func InitializeReportHandler(
	service entity_finance.FinancialReportDataServiceInterface,
	encryptData cryptdata.CryptDataInterface,
	authClient authenticatior.Authenticator,
	routerGroup *gin.RouterGroup,
	middleware ...gin.HandlerFunc,
) *ReportHandler {

	handler := &ReportHandler{
		service:     service,
		encryptData: encryptData,
		authClient:  authClient,
	}

	handler.setupRoutes(routerGroup, middleware...)
	return handler
}

// setupRoutes sets up the routes for spending plan operations under the given router group.
func (h *ReportHandler) setupRoutes(routerGroup *gin.RouterGroup, middleware ...gin.HandlerFunc) {

	report := routerGroup.Group("/finance/reports")
	for _, mw := range middleware {
		report.Use(mw)
	}

	report.GET("", h.GetReport)
	// spendingPlanGroup.PUT("", h.SaveSpendingPlan)
}

// GetSpendingPlan handles the GET /spending-plan request.
func (h *ReportHandler) GetReport(c *gin.Context) {
	log.Println("[HANDLER] start report >>>")
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	log.Println("[HANDLER] before service >>>")
	result, err := h.service.GetFinancialReportData(ctx)
	if err != nil {
		log.Println("[HANDLER] error after service >>>", err, result)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve income record: " + err.Error()})
		}
		return
	}

	log.Println("[HANDLER] after service >>>")
	log.Println("[HANDLER] >>>", result)

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
