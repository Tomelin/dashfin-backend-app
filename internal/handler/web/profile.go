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

	var payload cryptdata.CryptData

	// dados recebidos e realizado o bind para CrypstData
	// Essa parte está funcionando 100%
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// payloadData recebe o valor de payload, como string para fazer o Decript
	// Essa parte está funcionando 100%
	data, err := cryptdata.PayloadData(payload.Payload)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println("data received...", string(data))
	// Executa o bind o []byte recebido, para a struct Profile
	// Essa parte está funcionando 100%
	var profile entity_profile.Profile
	err = json.Unmarshal(data, &profile)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := json.Marshal(profile)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Executa o encript do payload
	// Essa parte está funcionando 100%
	result, err := cryptdata.EncryptPayload(b)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Executa o Decript do result, para validar se o frontend irá conseguir fazer o mesmo
	// Retorna esse error:  decryption failed: decrypt: failed to unpad data: pkcs7Unpad: invalid padding length (possible wrong key or corrupted data)
	// Mas como retorna erro, se nós que estamos gerando com a mesma chave do Encrypt
	data, err = cryptdata.PayloadData(result)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Printa os dados do valor recebido pelo payload e pelo o profile que geramos
	// Não estamos chegando essa etapa
	log.Println("data hanlder...", string(data))

	c.JSON(http.StatusOK, gin.H{"payload": result})
}
