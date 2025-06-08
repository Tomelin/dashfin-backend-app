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

type BankAccountHandlerInterface interface {
	CreateBankAccount(c *gin.Context)
	GetBankAccount(c *gin.Context)
	GetBankAccounts(c *gin.Context)
	UpdateBankAccount(c *gin.Context)
	DeleteBankAccount(c *gin.Context)
}

type BankAccountHandler struct {
	service     entity_finance.BankAccountServiceInterface
	router      *gin.RouterGroup
	encryptData cryptdata.CryptDataInterface
	authClient  authenticatior.Authenticator
}

func InitializeBankAccountHandler(svc entity_finance.BankAccountServiceInterface, encryptData cryptdata.CryptDataInterface, authClient authenticatior.Authenticator, routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) BankAccountHandlerInterface {
	handler := &BankAccountHandler{
		service:     svc,
		router:      routerGroup,
		encryptData: encryptData,
		authClient:  authClient,
	}

	handler.setupRoutes(routerGroup, middleware...)

	return handler
}

func (h *BankAccountHandler) setupRoutes(routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) {
	middlewareList := make([]gin.HandlerFunc, len(middleware))
	for i, mw := range middleware {
		middlewareList[i] = mw
	}

	routerGroup.POST("/finance/bank-accounts", append(middlewareList, h.CreateBankAccount)...)
	routerGroup.GET("/finance/bank-accounts/:id", append(middlewareList, h.GetBankAccount)...)
	routerGroup.GET("/finance/bank-accounts", append(middlewareList, h.GetBankAccounts)...)
	routerGroup.PUT("/finance/bank-accounts/:id", append(middlewareList, h.UpdateBankAccount)...)
	routerGroup.DELETE("/finance/bank-accounts/:id", append(middlewareList, h.DeleteBankAccount)...)
}

func (h *BankAccountHandler) CreateBankAccount(c *gin.Context) {
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

	var bankAccount entity_finance.BankAccount
	err = json.Unmarshal(data, &bankAccount)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println(bankAccount)
	log.Println(userId)
	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userId)

	result, err := h.service.CreateBankAccount(ctx, &bankAccount)
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

func (h *BankAccountHandler) GetBankAccount(c *gin.Context) {
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

	result, err := h.service.GetBankAccountByID(ctx, &id)
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

func (h *BankAccountHandler) GetBankAccounts(c *gin.Context) {
	userId, token, err := web.GetRequiredHeaders(h.authClient, c.Request)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userId)

	results, err := h.service.GetBankAccounts(ctx)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
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

func (h *BankAccountHandler) UpdateBankAccount(c *gin.Context) {
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

	var bankAccount entity_finance.BankAccountRequest
	err = json.Unmarshal(data, &bankAccount)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("getHunmarshaleader %", err.Error())})
		return
	}

	log.Println("ID", bankAccount.ID)
	log.Println(userId)
	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	ctx = context.WithValue(ctx, "UserID", userId)

	result, err := h.service.UpdateBankAccount(ctx, &bankAccount)
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

func (h *BankAccountHandler) DeleteBankAccount(c *gin.Context) {
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

	err = h.service.DeleteBankAccount(ctx, &id)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
