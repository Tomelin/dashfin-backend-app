package web

import (
	"context"
	"encoding/json"
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
	GetProfile(c *gin.Context)
	UpdateProfessional(c *gin.Context)
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

	routerGroup.GET("/profile/personal", append(middlewareList, cat.GetProfile)...)
	routerGroup.PUT("/profile/personal", append(middlewareList, cat.UpdateProfile)...)
	routerGroup.OPTIONS("/profile/personal", append(middlewareList, cat.Personal)...)

	routerGroup.PUT("/profile/professional", append(middlewareList, cat.UpdateProfessional)...)
	routerGroup.OPTIONS("/profile/professional", append(middlewareList, cat.UpdateProfessional)...)
	routerGroup.GET("/profile/professional", append(middlewareList, cat.GetProfessional)...)

	routerGroup.PUT("/profile/goals", append(middlewareList, cat.UpdateGoals)...)
	routerGroup.OPTIONS("/profile/goals", append(middlewareList, cat.UpdateGoals)...)
	routerGroup.GET("/profile/goals", append(middlewareList, cat.GetGoals)...)
}

func (cat *ProfileHandlerHttp) Create(c *gin.Context) {
	c.JSON(http.StatusOK, "created")
}

func (cat *ProfileHandlerHttp) Personal(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"payload": "ok"})
}

func (cat *ProfileHandlerHttp) UpdateProfile(c *gin.Context) {

	// Valida o header
	userId, token, err := GetRequiredHeaders(cat.authClient, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// bind crypt payload
	var payload cryptdata.CryptData
	err = c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//decrypt payload
	data, err := cat.encryptData.PayloadData(payload.Payload)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// bind profile
	var profile entity_profile.Profile
	err = json.Unmarshal(data, &profile)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile.UserProviderID = userId

	// update Profile
	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	user, err := cat.service.UpdateProfile(ctx, &profile)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := json.Marshal(user)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := cat.encryptData.EncryptPayload(b)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": result})
}

func (cat *ProfileHandlerHttp) GetProfile(c *gin.Context) {

	// Valida o header
	userId, token, err := GetRequiredHeaders(cat.authClient, c.Request)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// get profile
	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	user, err := cat.service.GetProfileByID(ctx, &userId)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := json.Marshal(user)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := cat.encryptData.EncryptPayload(b)
	if err != nil {

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

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = cat.encryptData.PayloadData(payload.Payload)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": "ok"})
}

func (cat *ProfileHandlerHttp) UpdateProfessional(c *gin.Context) {

	// Valida o header
	userId, token, err := GetRequiredHeaders(cat.authClient, c.Request)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// bind crypt payload
	var payload cryptdata.CryptData
	err = c.ShouldBindJSON(&payload)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//decrypt payload
	data, err := cat.encryptData.PayloadData(payload.Payload)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// bind profile
	var profession entity_profile.ProfileProfession
	err = json.Unmarshal(data, &profession)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// update Profile
	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	professionResult, err := cat.service.UpdateProfileProfession(ctx, &userId, &profession)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := json.Marshal(professionResult)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := cat.encryptData.EncryptPayload(b)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": result})
}

func (cat *ProfileHandlerHttp) GetProfessional(c *gin.Context) {

	// Valida o header
	userId, token, err := GetRequiredHeaders(cat.authClient, c.Request)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// get profile
	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	profession, err := cat.service.GetProfileProfession(ctx, &userId)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := json.Marshal(profession)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := cat.encryptData.EncryptPayload(b)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": result})
}

func (cat *ProfileHandlerHttp) UpdateGoals(c *gin.Context) {

	// Valida o header
	userId, token, err := GetRequiredHeaders(cat.authClient, c.Request)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// bind crypt payload
	var payload cryptdata.CryptData
	err = c.ShouldBindJSON(&payload)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//decrypt payload
	data, err := cat.encryptData.PayloadData(payload.Payload)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// bind profile
	var goals entity_profile.ProfileGoals
	err = json.Unmarshal(data, &goals)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// update Profile
	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	professionResult, err := cat.service.UpdateProfileGoals(ctx, &userId, &goals)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := json.Marshal(professionResult)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := cat.encryptData.EncryptPayload(b)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": result})
}

func (cat *ProfileHandlerHttp) GetGoals(c *gin.Context) {

	// Valida o header
	userId, token, err := GetRequiredHeaders(cat.authClient, c.Request)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// get profile
	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	profession, err := cat.service.GetProfileGoals(ctx, &userId)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := json.Marshal(profession)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := cat.encryptData.EncryptPayload(b)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": result})
}
