package web

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Tomelin/dashfin-backend-app/internal/core/entity"
	"github.com/Tomelin/dashfin-backend-app/internal/core/service"
	"github.com/Tomelin/dashfin-backend-app/pkg/authenticatior"
	cryptdata "github.com/Tomelin/dashfin-backend-app/pkg/cryptData"
	"github.com/gin-gonic/gin"
)

type SupportHandlerHttpInterface interface {
	CreateSupport(c *gin.Context)
}

type SupportHandlerHttp struct {
	service     service.SupportServiceInterface
	router      *gin.RouterGroup
	encryptData cryptdata.CryptDataInterface
	authClient  authenticatior.Authenticator
}

func InicializationSupportHandlerHttp(svc service.SupportServiceInterface, encryptData cryptdata.CryptDataInterface, authClient authenticatior.Authenticator, routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) SupportHandlerHttpInterface {

	load := &SupportHandlerHttp{
		service:     svc,
		router:      routerGroup,
		encryptData: encryptData,
		authClient:  authClient,
	}

	load.handlers(routerGroup, middleware...)

	return load
}

func (cat *SupportHandlerHttp) handlers(routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) {
	middlewareList := make([]gin.HandlerFunc, len(middleware))
	for i, mw := range middleware {
		middlewareList[i] = mw
	}

	routerGroup.POST("/support/requests", append(middlewareList, cat.CreateSupport)...)
	routerGroup.OPTIONS("/support/requests", append(middlewareList, cat.CreateSupport)...)

}

func (cat *SupportHandlerHttp) CreateSupport(c *gin.Context) {
	// Valida o header
	userId, token, err := getRequiredHeaders(cat.authClient, c.Request)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// bind crypt payload
	var payload cryptdata.CryptData
	err = c.ShouldBindJSON(&payload)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//decrypt payload
	data, err := cat.encryptData.PayloadData(payload.Payload)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// bind support
	var support entity.Support
	err = json.Unmarshal(data, &support)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	support.UserProviderID = userId

	// update support
	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	supportResponse, err := cat.service.Create(ctx, &support)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := json.Marshal(supportResponse)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := cat.encryptData.EncryptPayload(b)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": result})
}
