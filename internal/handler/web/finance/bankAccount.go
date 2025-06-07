package finance

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	service_finance "github.com/Tomelin/dashfin-backend-app/internal/core/service/finance"
	"github.com/Tomelin/dashfin-backend-app/internal/handler/web"

	"github.com/Tomelin/dashfin-backend-app/pkg/authenticatior"
	cryptdata "github.com/Tomelin/dashfin-backend-app/pkg/cryptData"
	"github.com/gin-gonic/gin"
)

type BankAccountHandlerHttpInterface interface {
	CreateBankAccount(c *gin.Context)
	GetBankAccountByID(c *gin.Context)
	GetBankAccounts(c *gin.Context)
	UpdateBankAccount(c *gin.Context)
	DeleteBankAccount(c *gin.Context)
}

type BankAccountHandlerHttp struct {
	service     service_finance.BankAccountServiceInterface
	router      *gin.RouterGroup
	encryptData cryptdata.CryptDataInterface
	authClient  authenticatior.Authenticator
}

func InitializeBankAccountHandlerHttp(svc service_finance.BankAccountServiceInterface, encryptData cryptdata.CryptDataInterface, authClient authenticatior.Authenticator, routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) BankAccountHandlerHttpInterface {

	load := &BankAccountHandlerHttp{
		service:     svc,
		router:      routerGroup,
		encryptData: encryptData,
		authClient:  authClient,
	}

	load.handlers(routerGroup, middleware...)

	return load
}

func (bah *BankAccountHandlerHttp) handlers(routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) {
	middlewareList := make([]gin.HandlerFunc, len(middleware))
	for i, mw := range middleware {
		middlewareList[i] = mw
	}

	routerGroup.POST("/finance/bank-accounts", append(middlewareList, bah.CreateBankAccount)...)
	routerGroup.GET("/finance/bank-accounts/:id", append(middlewareList, bah.GetBankAccountByID)...)
	routerGroup.GET("/finance/bank-accounts", append(middlewareList, bah.GetBankAccounts)...)
	routerGroup.PUT("/finance/bank-accounts/:id", append(middlewareList, bah.UpdateBankAccount)...)
	routerGroup.DELETE("/finance/bank-accounts/:id", append(middlewareList, bah.DeleteBankAccount)...)
	routerGroup.OPTIONS("/finance/bank-accounts", append(middlewareList, bah.optionsHandler)...)
	routerGroup.OPTIONS("/finance/bank-accounts/:id", append(middlewareList, bah.optionsHandler)...)

	routerGroup.POST("/lookups/financial-institutions", append(middlewareList, bah.CreateBankAccount)...)
	routerGroup.GET("/lookups/financial-institutions/:id", append(middlewareList, bah.GetBankAccountByID)...)
	routerGroup.GET("/lookups/financial-institutions", append(middlewareList, bah.GetBankAccounts)...)
	routerGroup.PUT("/lookups/financial-institutions/:id", append(middlewareList, bah.UpdateBankAccount)...)
	routerGroup.DELETE(" /lookups/financial-institutions/:id", append(middlewareList, bah.DeleteBankAccount)...)
	routerGroup.OPTIONS("/lookups/financial-institutions", append(middlewareList, bah.optionsHandler)...)
	routerGroup.OPTIONS("/lookups/financial-institutions/:id", append(middlewareList, bah.optionsHandler)...)
}

func (bah *BankAccountHandlerHttp) optionsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"payload": "ok"})
}

func (bah *BankAccountHandlerHttp) CreateBankAccount(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(bah.authClient, c.Request)
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

	data, err := bah.encryptData.PayloadData(payload.Payload)
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

	ctx := context.WithValue(c.Request.Context(), web.AuthTokenKey, token)
	result, err := bah.service.CreateBankAccount(ctx, userID, &bankAccount)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := json.Marshal(result)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	encryptedResult, err := bah.encryptData.EncryptPayload(b)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"payload": encryptedResult})
}

func (bah *BankAccountHandlerHttp) GetBankAccountByID(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(bah.authClient, c.Request)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bank account ID is required"})
		return
	}

	ctx := context.WithValue(c.Request.Context(), web.AuthTokenKey, token)
	result, err := bah.service.GetBankAccountByID(ctx, userID, id)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := json.Marshal(result)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	encryptedResult, err := bah.encryptData.EncryptPayload(b)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": encryptedResult})
}

func (bah *BankAccountHandlerHttp) GetBankAccounts(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(bah.authClient, c.Request)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.WithValue(c.Request.Context(), web.AuthTokenKey, token)
	result, err := bah.service.GetBankAccounts(ctx, userID)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := json.Marshal(result)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	encryptedResult, err := bah.encryptData.EncryptPayload(b)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": encryptedResult})
}

func (bah *BankAccountHandlerHttp) UpdateBankAccount(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(bah.authClient, c.Request)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bank account ID is required"})
		return
	}

	var payload cryptdata.CryptData
	err = c.ShouldBindJSON(&payload)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := bah.encryptData.PayloadData(payload.Payload)
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

	ctx := context.WithValue(c.Request.Context(), web.AuthTokenKey, token)
	result, err := bah.service.UpdateBankAccount(ctx, userID, id, &bankAccount)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := json.Marshal(result)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	encryptedResult, err := bah.encryptData.EncryptPayload(b)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": encryptedResult})
}

func (bah *BankAccountHandlerHttp) DeleteBankAccount(c *gin.Context) {
	userID, token, err := web.GetRequiredHeaders(bah.authClient, c.Request)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bank account ID is required"})
		return
	}

	ctx := context.WithValue(c.Request.Context(), web.AuthTokenKey, token)
	err = bah.service.DeleteBankAccount(ctx, userID, id)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}
