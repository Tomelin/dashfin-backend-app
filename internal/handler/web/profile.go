package web

import (
	"encoding/json"
	"log"
	"net/http"

	entity_profile "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
	cryptdata "github.com/Tomelin/dashfin-backend-app/pkg/cryptData"
	"github.com/gin-gonic/gin"
)

type HandlerHttpInterface interface {
	Create(c *gin.Context)
	Personal(c *gin.Context)
	PutPersonal(c *gin.Context)
	// Get(c *gin.Context)
	// GetById(c *gin.Context)
	// Update(c *gin.Context)
	// Delete(c *gin.Context)
	// GetByFilterMany(c *gin.Context)
	// GetByFilterOne(c *gin.Context)
}

type ProfileHandlerHttp struct {
	Service     string
	router      *gin.RouterGroup
	encryptData cryptdata.CryptDataInterface
}

func InicializationProfileHandlerHttp(svc string, encryptData cryptdata.CryptDataInterface, routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) HandlerHttpInterface {

	load := &ProfileHandlerHttp{
		Service:     svc,
		router:      routerGroup,
		encryptData: encryptData,
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
	routerGroup.PUT("/profile/personal", append(middlewareList, cat.PutPersonal)...)
	routerGroup.GET("/profile/personal", append(middlewareList, cat.Personal)...)
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
	log.Println("get new request")

	var prof entity_profile.Profile

	err := c.ShouldBindJSON(&prof)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println(prof)
	log.Println("after bind json")

	profile := entity_profile.Profile{
		FullName:  "name of client",
		Phone:     "51984104084",
		BirthDate: "20/05/2025",
		Email:     "email@teste.com.br",
	}

	log.Println(profile)
	c.JSON(http.StatusOK, profile)
}

func (cat *ProfileHandlerHttp) PutPersonal(c *gin.Context) {
	log.Println("  PutPersonal > new request")

	var payload cryptdata.CryptData

	err := c.ShouldBindJSON(&payload)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("pyload was bind....", payload.Payload)

	cryptdata.PayloadData(payload.Payload)
	cat.encryptData.GetToken()

	data, err := cat.encryptData.DecodePayload(&payload.Payload)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println("after bind json")

	var prof entity_profile.Profile

	err = json.Unmarshal(data, &prof)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println(prof)

	err = c.ShouldBindJSON(&prof)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println(prof)
	log.Println("after bind json")

	// profile := entity_profile.Profile{
	// 	FullName:  "name of client",
	// 	Phone:     "51984104084",
	// 	BirthDate: "20/05/2025",
	// 	Email:     "email@teste.com.br",
	// }

	c.JSON(http.StatusOK, payload.Payload)
}
