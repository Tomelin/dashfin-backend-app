package web_finance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	web "github.com/Tomelin/dashfin-backend-app/internal/handler/web"
	"github.com/Tomelin/dashfin-backend-app/pkg/authenticatior"
	cryptdata "github.com/Tomelin/dashfin-backend-app/pkg/cryptData"
	"github.com/gin-gonic/gin"
)

type CreditCardHandlerInterface interface {
	CreateCreditCard(c *gin.Context)
	GetCreditCard(c *gin.Context)
	GetCreditCards(c *gin.Context)
	UpdateCreditCard(c *gin.Context)
	DeleteCreditCard(c *gin.Context)
}

type CreditCardHandler struct {
	service     entity_finance.CreditCardServiceInterface
	router      *gin.RouterGroup
	encryptData cryptdata.CryptDataInterface
	authClient  authenticatior.Authenticator
}

func InitializeCreditCardHandler(svc entity_finance.CreditCardServiceInterface, encryptData cryptdata.CryptDataInterface, authClient authenticatior.Authenticator, routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) CreditCardHandlerInterface {
	handler := &CreditCardHandler{
		service:     svc,
		router:      routerGroup,
		encryptData: encryptData,
		authClient:  authClient,
	}

	handler.setupRoutes(routerGroup, middleware...)

	return handler
}

func (h *CreditCardHandler) setupRoutes(routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) {
	middlewareList := make([]gin.HandlerFunc, len(middleware))
	for i, mw := range middleware {
		middlewareList[i] = mw
	}

	routerGroup.POST("/finance/credit-cards", append(middlewareList, h.CreateCreditCard)...)
	routerGroup.GET("/finance/credit-cards/:id", append(middlewareList, h.GetCreditCard)...)
	routerGroup.GET("/finance/credit-cards", append(middlewareList, h.GetCreditCards)...)
	routerGroup.PUT("/finance/credit-cards/:id", append(middlewareList, h.UpdateCreditCard)...)
	routerGroup.DELETE("/finance/credit-cards/:id", append(middlewareList, h.DeleteCreditCard)...)
}

func (h *CreditCardHandler) CreateCreditCard(c *gin.Context) {
	userId, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var payload cryptdata.CryptData
	err = c.ShouldBindJSON(&payload)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := h.encryptData.PayloadData(payload.Payload)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var creditCard entity_finance.CreditCard
	err = json.Unmarshal(data, &creditCard)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userId)

	result, err := h.service.CreateCreditCard(ctx, &creditCard)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := json.Marshal(result)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	encryptedResult, err := h.encryptData.EncryptPayload(b)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"payload": encryptedResult})
}

func (h *CreditCardHandler) GetCreditCard(c *gin.Context) {
	userId, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userId)

	result, err := h.service.GetCreditCardByID(ctx, &id)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	b, err := json.Marshal(result)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	encryptedResult, err := h.encryptData.EncryptPayload(b)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": encryptedResult})
}

func (h *CreditCardHandler) GetCreditCards(c *gin.Context) {
	userId, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userId)

	results, err := h.service.GetCreditCards(ctx)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	for _, v := range results {
		log.Println("ID", v.ID, "object", v)
	}

	b, err := json.Marshal(results)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	encryptedResult, err := h.encryptData.EncryptPayload(b)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": encryptedResult})
}

func (h *CreditCardHandler) UpdateCreditCard(c *gin.Context) {
	userId, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("getHeader %", err.Error())})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	var payload cryptdata.CryptData
	err = c.ShouldBindJSON(&payload)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("cryptData %", err.Error())})
		return
	}

	data, err := h.encryptData.PayloadData(payload.Payload)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("payloadData %", err.Error())})
		return
	}

	var creditCard entity_finance.CreditCardRequest
	err = json.Unmarshal(data, &creditCard)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("getHunmarshaleader %", err.Error())})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userId)

	result, err := h.service.UpdateCreditCard(ctx, &creditCard)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Update %", err.Error())})
		return
	}

	b, err := json.Marshal(result)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Marshal %", err.Error())})
		return
	}

	encryptedResult, err := h.encryptData.EncryptPayload(b)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("encrypt %", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": encryptedResult})
}

func (h *CreditCardHandler) DeleteCreditCard(c *gin.Context) {
	userId, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userId)

	err = h.service.DeleteCreditCard(ctx, &id)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}