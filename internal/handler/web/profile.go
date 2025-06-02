package web

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HandlerHttpInterface interface {
	Create(c *gin.Context)
	Personal(c *gin.Context)
	// Get(c *gin.Context)
	// GetById(c *gin.Context)
	// Update(c *gin.Context)
	// Delete(c *gin.Context)
	// GetByFilterMany(c *gin.Context)
	// GetByFilterOne(c *gin.Context)
}

type ProfileHandlerHttp struct {
	Service string
	router  *gin.RouterGroup
}

func InicializationProfileHandlerHttp(svc string, routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) HandlerHttpInterface {

	load := &ProfileHandlerHttp{
		Service: svc,
		router:  routerGroup,
	}

	load.handlers(routerGroup, middleware...)

	return load
}

func (cat *ProfileHandlerHttp) handlers(routerGroup *gin.RouterGroup, middleware ...func(c *gin.Context)) {
	middlewareList := make([]gin.HandlerFunc, len(middleware))
	for i, mw := range middleware {
		middlewareList[i] = mw
	}

	routerGroup.POST("/profile", append(middlewareList, cat.Create)...)
	routerGroup.GET("/profile/", append(middlewareList, cat.Create)...)
	routerGroup.GET("/profile/:id", append(middlewareList, cat.Create)...)
	routerGroup.PUT("/profile/:id", append(middlewareList, cat.Create)...)
	routerGroup.PUT("/profile/personal", append(middlewareList, cat.Personal)...)
	routerGroup.GET("/profile/personal", append(middlewareList, cat.Personal)...)
	routerGroup.POST("/profile/personal", append(middlewareList, cat.Personal)...)
	routerGroup.DELETE("/profile/:id", append(middlewareList, cat.Create)...)
	routerGroup.GET("/profile/search", append(middlewareList, cat.Create)...)
	routerGroup.GET("/profile/filter", append(middlewareList, cat.Create)...)
}

func (cat *ProfileHandlerHttp) Create(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}

func (cat *ProfileHandlerHttp) Personal(c *gin.Context) {
	log.Println("ok")
	c.JSON(http.StatusOK, "ok")
}
