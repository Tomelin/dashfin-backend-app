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

	log.Println("recebendo o options....")
	c.Header("Access-Control-Allow-Origin", "*") // Ou sua origem do frontend
	c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, X-Authorization, X-USERID, X-APP, X-TRACE-ID")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Content-Type", "application/json")

	c.Status(http.StatusNoContent)
}

func (cat *SupportHandlerHttp) CreateSupport(c *gin.Context) {
	log.Println("Backend: CreateSupport handler iniciado para POST /api/support/requests") // Adicionar log

	log.Println("step 1")
	// 1. Valida o header
	userId, token, err := getRequiredHeaders(cat.authClient, c.Request)
	if err != nil {
		log.Printf("Backend ERROR (CreateSupport - getRequiredHeaders): %v\n", err) // Log mais detalhado
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro de autenticação: " + err.Error()})
		return
	}
	log.Printf("Backend INFO (CreateSupport): Headers validados. UserID: %s\n", userId)

	// 2. bind crypt payload
	var payload cryptdata.CryptData
	// Logar o corpo bruto antes do bind pode ser útil se o bind falhar
	// rawBody, _ := io.ReadAll(c.Request.Body)
	// log.Printf("Backend DEBUG (CreateSupport): Corpo da requisição bruto recebido: %s\n", string(rawBody))
	// c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody)) // Restaurar o corpo para o bind

	err = c.ShouldBindJSON(&payload)
	if err != nil {
		log.Printf("Backend ERROR (CreateSupport - ShouldBindJSON): %v\n", err) // Log mais detalhado
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar payload da requisição: " + err.Error()})
		return
	}
	log.Println("step 3")
	if payload.Payload == "" {
		log.Println("Backend ERROR (CreateSupport): Campo 'payload' está vazio no JSON recebido.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload criptografado ausente ou inválido."})
		return
	}
	log.Printf("Backend INFO (CreateSupport): Payload JSON vinculado. Payload Base64 (prefixo): %s...\n", payload.Payload[:min(len(payload.Payload), 50)])

	// 3. decrypt payload
	data, err := cat.encryptData.PayloadData(payload.Payload)
	if err != nil {
		log.Printf("Backend ERROR (CreateSupport - PayloadData decrypt): %v\n", err) // Log mais detalhado
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao descriptografar dados: " + err.Error()})
		return
	}
	log.Printf("Backend INFO (CreateSupport): Dados descriptografados (prefixo): %s...\n", string(data)[:min(len(string(data)), 100)])

	// 4. bind support entity
	var support entity.Support // Verifique se os campos em entity.Support correspondem ao que o frontend envia (category, description)
	err = json.Unmarshal(data, &support)
	if err != nil {
		log.Printf("Backend ERROR (CreateSupport - json.Unmarshal para entity.Support): %v\n", err)
		log.Printf("Backend DEBUG (CreateSupport): Dados JSON que falharam no Unmarshal: %s\n", string(data)) // Log crucial
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao desserializar dados do suporte: " + err.Error()})
		return
	}
	log.Printf("Backend INFO (CreateSupport): Dados desserializados para entity.Support: %+v\n", support)

	support.UserProviderID = userId

	// 5. Chamar o serviço
	ctx := context.WithValue(c.Request.Context(), "Authorization", token)
	supportResponse, err := cat.service.Create(ctx, &support)
	if err != nil {
		log.Printf("Backend ERROR (CreateSupport - service.Create): %v\n", err)                                                // Log mais detalhado
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao criar solicitação de suporte: " + err.Error()}) // Use 500 para erros de serviço
		return
	}
	log.Printf("Backend INFO (CreateSupport): Serviço de suporte executado. Resposta do serviço: %+v\n", supportResponse)

	// 6. Marshal da resposta
	b, err := json.Marshal(supportResponse)
	if err != nil {
		log.Printf("Backend ERROR (CreateSupport - json.Marshal da resposta): %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao serializar resposta: " + err.Error()})
		return
	}
	log.Printf("Backend INFO (CreateSupport): Resposta serializada (prefixo): %s...\n", string(b)[:min(len(string(b)), 100)])

	// 7. Criptografar resposta
	result, err := cat.encryptData.EncryptPayload(b)
	if err != nil {
		log.Printf("Backend ERROR (CreateSupport - EncryptPayload da resposta): %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criptografar resposta: " + err.Error()})
		return
	}
	log.Printf("Backend INFO (CreateSupport): Resposta criptografada (prefixo Base64): %s...\n", result[:min(len(result), 50)])

	c.JSON(http.StatusOK, gin.H{"payload": result})
	log.Println("Backend INFO (CreateSupport): Resposta 200 OK enviada com payload criptografado.")
}

// Adicione esta função auxiliar se não existir no seu pacote
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
