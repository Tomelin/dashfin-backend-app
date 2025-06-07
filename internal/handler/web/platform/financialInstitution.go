package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	service "github.com/Tomelin/dashfin-backend-app/internal/core/entity/platform"
	web_types "github.com/Tomelin/dashfin-backend-app/internal/handler/web"
	"github.com/Tomelin/dashfin-backend-app/pkg/authenticatior"
	cryptdata "github.com/Tomelin/dashfin-backend-app/pkg/cryptData"
)

type FinancialInstitutionHandler struct {
	router      *gin.RouterGroup
	service     service.FinancialInstitutionInterface
	encryptData cryptdata.CryptDataInterface
	authClient  authenticatior.Authenticator
}

func NewFinancialInstitutionHandler(svc service.FinancialInstitutionInterface, encryptData cryptdata.CryptDataInterface, authClient authenticatior.Authenticator, routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) FinancialInstitutionHandler {
	load := FinancialInstitutionHandler{
		service:     svc,
		router:      routerGroup,
		encryptData: encryptData,
		authClient:  authClient,
	}

	load.setupRoutes(routerGroup, middleware...)

	return load
}

func (h *FinancialInstitutionHandler) setupRoutes(routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) {
	middlewareList := make([]gin.HandlerFunc, len(middleware))
	for i, mw := range middleware {
		middlewareList[i] = mw
	}

	financialInstitutionsGroup := h.router.Group("/lookups/financial-institutions")
	{
		financialInstitutionsGroup.Use(middlewareList...)
		financialInstitutionsGroup.GET("", h.getFinancialInstitutions)
		financialInstitutionsGroup.GET("/:id", h.getFinancialInstitution)
	}
}

func (cat *FinancialInstitutionHandler) HandleSupportOptions(c *gin.Context) {

	c.Header("Access-Control-Allow-Origin", "*") // Ou sua origem do frontend
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, X-Authorization, X-USERID, X-APP, X-TRACE-ID")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Content-Type", "application/json")

	c.Status(http.StatusNoContent)
}

// @Summary Get all financial institutions
// @Description Retrieve a list of all financial institutions
// @Tags Financial Institutions
// @Produce json
// @Success 200 {array} response.FinancialInstitutionResponse "List of financial institutions"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/financial-institutions [get]
func (h *FinancialInstitutionHandler) getFinancialInstitutions(c *gin.Context) {

	// Header validate
	_, token, err := web_types.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro de autenticação: " + err.Error()})
		return
	}

	// Get user token
	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	institutions, err := h.service.GetAllFinancialInstitutions(ctx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payloadEnconded, err := json.Marshal(institutions)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payloadEncrypted, err := h.encryptData.EncryptPayload(payloadEnconded)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": payloadEncrypted})
}

// @Summary Get a financial institution by ID
// @Description Retrieve a financial institution by its unique identifier
// @Tags Financial Institutions
// @Produce json
// @Param id path string true "Financial Institution ID"
// @Success 200 {object} response.FinancialInstitutionResponse "Financial institution details"
// @Failure 400 {object} response.ErrorResponse "Invalid financial institution ID"
// @Failure 404 {object} response.ErrorResponse "Financial institution not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/financial-institutions/{id} [get]
func (h *FinancialInstitutionHandler) getFinancialInstitution(c *gin.Context) {
	idParam := c.Param("id")

	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid financial institution ID"})
		return
	}

	// Header validate
	_, token, err := web_types.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro de autenticação: " + err.Error()})
		return
	}

	// Get user token
	ctx := context.WithValue(c.Request.Context(), web_types.AuthTokenKey, token)
	institution, err := h.service.GetFinancialInstitutionByID(ctx, &idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payloadEnconded, err := json.Marshal(institution)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payloadEncrypted, err := h.encryptData.EncryptPayload(payloadEnconded)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": payloadEncrypted})
}
