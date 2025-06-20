package web_dashboard

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	dashboardEntity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/dashboard"
	"github.com/Tomelin/dashfin-backend-app/internal/handler/web"
	"github.com/Tomelin/dashfin-backend-app/pkg/authenticatior"
	cryptdata "github.com/Tomelin/dashfin-backend-app/pkg/cryptData"
	// Service import will be used by the initializer that wires up the actual service
	// "github.com/Tomelin/dashfin-backend-app/internal/core/service/finance"
)

// DashboardServiceInterface defines the methods our handler expects from the dashboard service.
type DashboardServiceInterface interface {
	GetDashboardData(ctx context.Context) (*dashboardEntity.Dashboard, error)
}

// DashboardHandler handles HTTP requests for dashboard data using Gin.
type DashboardHandler struct {
	service     DashboardServiceInterface
	encryptData cryptdata.CryptDataInterface
	authClient  authenticatior.Authenticator
	// router  *gin.RouterGroup // Not strictly needed if setupRoutes is called by initializer
}

// InitializeDashboardHandler sets up the dashboard handler with its routes.
// This function would be called from the main application setup.
// It mirrors the pattern seen in BankAccountHandler.
func InitializeDashboardHandler(
	svc DashboardServiceInterface,
	encryptData cryptdata.CryptDataInterface,
	authClient authenticatior.Authenticator,
	routerGroup *gin.RouterGroup,
	middleware ...gin.HandlerFunc,
) *DashboardHandler {
	handler := &DashboardHandler{
		service:     svc,
		encryptData: encryptData,
		authClient:  authClient,
	}

	handler.setupRoutes(routerGroup, middleware...)
	return handler
}

func (h *DashboardHandler) setupRoutes(
	routerGroup *gin.RouterGroup,
	// authMiddleware gin.HandlerFunc,
	middleware ...gin.HandlerFunc,
) {

	// All middleware passed to InitializeDashboardHandler will be applied.
	dashboardRoutes := routerGroup.Group("/dashboard/summary")
	for _, mw := range middleware {
		dashboardRoutes.Use(mw)
	}

	dashboardRoutes.GET("", append(middleware, h.GetDashboard)...)
}

// GetDashboard is the Gin HTTP handler for GET /dashboard requests.
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userID)

	results, err := h.service.GetDashboardData(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve dashboard data"})
		return
	}

	if results == nil {
		c.JSON(http.StatusNoContent, gin.H{"message": "there are contents"})
		return
	}

	bills := make([]dashboardEntity.UpcomingBillData, 0)

	for _, v := range results.UpcomingBillsData {
		bills = append(bills, dashboardEntity.UpcomingBillData{
			Name:    v.BillName,
			Amount:  v.Amount,
			DueDate: v.DueDate,
		})
	}

	results.SummaryCards.UpcomingBillsData = bills

	responseBytes, err := json.Marshal(results.SummaryCards)
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
