package web

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	entity_profile "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
	service "github.com/Tomelin/dashfin-backend-app/internal/core/service/profile"
	"github.com/Tomelin/dashfin-backend-app/pkg/authenticatior"
	cryptdata "github.com/Tomelin/dashfin-backend-app/pkg/cryptData"
	"github.com/gin-gonic/gin"
)

type HandlerHttpInterface interface {
	Create(c *gin.Context)
	Personal(c *gin.Context)
	UpdateProfile(c *gin.Context)
	// Get(c *gin.Context)
	// GetById(c *gin.Context)
	// Update(c *gin.Context)
	// Delete(c *gin.Context)
	// GetByFilterMany(c *gin.Context)
	// GetByFilterOne(c *gin.Context)
}

type ProfileHandlerHttp struct {
	service     service.ProfileServiceInterface
	router      *gin.RouterGroup
	encryptData cryptdata.CryptDataInterface
	authClient  authenticatior.Authenticator
}

func InicializationProfileHandlerHttp(svc service.ProfileServiceInterface, encryptData cryptdata.CryptDataInterface, authClient authenticatior.Authenticator, routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) HandlerHttpInterface {

	load := &ProfileHandlerHttp{
		service:     svc,
		router:      routerGroup,
		encryptData: encryptData,
		authClient:  authClient,
	}

	load.handlers(routerGroup, middleware...)

	return load
}

func (cat *ProfileHandlerHttp) handlers(routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) {
	middlewareList := make([]gin.HandlerFunc, len(middleware))
	for i, mw := range middleware {
		middlewareList[i] = mw
	}

	routerGroup.POST("/profile", append(middlewareList, cat.Personal)...)
	routerGroup.GET("/profile/", append(middlewareList, cat.Personal)...)
	routerGroup.GET("/profile/:id", append(middlewareList, cat.Personal)...)
	routerGroup.PUT("/profile/:id", append(middlewareList, cat.Personal)...)
	routerGroup.PUT("/profile/personal", append(middlewareList, cat.UpdateProfile)...)
	routerGroup.GET("/profile/personal", append(middlewareList, cat.Personal)...)
	routerGroup.POST("/profile/updateLogin", append(middlewareList, cat.UpdateLogin)...)
	routerGroup.POST("/profile/personal", append(middlewareList, cat.Personal)...)
	routerGroup.OPTIONS("/profile/personal", append(middlewareList, cat.Personal)...)
	routerGroup.DELETE("/profile/:id", append(middlewareList, cat.Personal)...)
	routerGroup.GET("/profile/search", append(middlewareList, cat.Personal)...)
	routerGroup.GET("/profile/filter", append(middlewareList, cat.Personal)...)
}

func (cat *ProfileHandlerHttp) Create(c *gin.Context) {
	c.JSON(http.StatusOK, "created")
}

func (cat *ProfileHandlerHttp) Personal(c *gin.Context) {

	log.Println(c.Request.Header)

	c.JSON(http.StatusOK, gin.H{"payload": "ok"})
}

func (cat *ProfileHandlerHttp) UpdateProfile(c *gin.Context) {

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

	// bind profile
	var profile entity_profile.Profile
	err = json.Unmarshal(data, &profile)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile.UserID = userId

	// update Profile
	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	user, err := cat.service.UpdateProfile(ctx, &profile)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := json.Marshal(user)
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

func (cat *ProfileHandlerHttp) UpdateLogin(c *gin.Context) {
	var payload cryptdata.CryptData

	// dados recebidos e realizado o bind para CrypstData
	// Essa parte está funcionando 100%
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println("received payload", payload)

	data, err := cat.encryptData.PayloadData(payload.Payload)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("received data", data)
	log.Println("received data string", string(data))
	c.JSON(http.StatusOK, gin.H{"payload": "ok"})
}
