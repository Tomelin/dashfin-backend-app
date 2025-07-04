package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Tomelin/dashfin-backend-app/internal/core/entity"
	"github.com/Tomelin/dashfin-backend-app/internal/core/service"
	"github.com/Tomelin/dashfin-backend-app/pkg/authenticatior"
	cryptdata "github.com/Tomelin/dashfin-backend-app/pkg/cryptData"
	"github.com/gin-gonic/gin"
)

type SupportHandlerHttpInterface interface {
	CreateSupport(c *gin.Context)
	HandleSupportOptions(c *gin.Context)
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
	routerGroup.OPTIONS("/support/requests", append(middlewareList, cat.HandleSupportOptions)...)
}

// Novo handler específico para OPTIONS
func (cat *SupportHandlerHttp) HandleSupportOptions(c *gin.Context) {

	c.Header("Access-Control-Allow-Origin", "*") // Ou sua origem do frontend
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type, X-Authorization, X-USERID, X-APP, X-TRACE-ID")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Content-Type", "application/json")

	c.Status(http.StatusNoContent)
}

func (cat *SupportHandlerHttp) CreateSupport(c *gin.Context) {

	// 1. Valida o header
	userId, token, err := GetRequiredHeaders(cat.authClient, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro de autenticação: " + err.Error()})
		return
	}

	// 2. bind crypt payload
	var payload cryptdata.CryptData
	// Logar o corpo bruto antes do bind pode ser útil se o bind falhar
	// rawBody, _ := io.ReadAll(c.Request.Body)
	// c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody)) // Restaurar o corpo para o bind

	err = c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar payload da requisição: " + err.Error()})
		return
	}

	if payload.Payload == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload criptografado ausente ou inválido."})
		return
	}

	// 3. decrypt payload
	data, err := cat.encryptData.PayloadData(payload.Payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao descriptografar dados: " + err.Error()})
		return
	}

	// 4. bind support entity
	var support entity.Support
	err = json.Unmarshal(data, &support)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao desserializar dados do suporte: " + err.Error()})
		return
	}

	support.UserProviderID = userId

	// 5. Chamar o serviço
	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	supportResponse, err := cat.service.Create(ctx, &support)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao criar solicitação de suporte: " + err.Error()}) // Use 500 para erros de serviço
		return
	}

	// 6. Marshal da resposta
	b, err := json.Marshal(supportResponse)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao serializar resposta: " + err.Error()})
		return
	}

	// 7. Criptografar resposta
	result, err := cat.encryptData.EncryptPayload(b)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criptografar resposta: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payload": result})
}

// Adicione esta função auxiliar se não existir no seu pacote
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
