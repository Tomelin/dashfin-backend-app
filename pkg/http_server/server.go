package http_server

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/http2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var cpuTemp = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "cpu_temperature_celsius",
	Help: "Current temperature of the CPU.",
})

var routerPath *gin.RouterGroup

var corsConfig = cors.Config{
	AllowOrigins:     []string{"http://localhost:3000", "https://studio--prosperar-n1en5.us-central1.hosted.app", "https://www.dashfin.com.br"}, // Adicione a origem do seu frontend
	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "X-AUTHORIZATION", "X-APP", "X-USERID", "X-TRACE-ID"},
	ExposeHeaders:    []string{"Content-Length"},
	AllowCredentials: true,
	MaxAge:           12 * time.Hour,
}

func init() {
	prometheus.MustRegister(cpuTemp)
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func (s *RestAPI) Run(handle http.Handler) error {

	s.Route.Use(s.CORSMiddleware)
	s.Route.Use(cors.New(corsConfig))
	s.Route.Use(s.CORSConfig)
	srv := http.Server{
		Addr:    ":" + s.Config.Port,
		Handler: s.Route.Handler(),
	}

	http2.ConfigureServer(&srv, &http2.Server{})
	s.Route.Use(s.ValidateToken)
	s.Route.Use(s.CORSMiddleware)
	s.Route.Use(s.CORSConfig)
	s.Route.Use(s.MiddlewareHeader)

	return srv.ListenAndServe()
}

func (s *RestAPI) RunTLS() error {
	return nil
}

func setHeader(c *gin.Context) {

	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers")
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")

	c.Next()
}

func (s *RestAPI) MiddlewareHeader(c *gin.Context) {
	const SkipMiddlewareKey = "skipMiddleware"

	if c.GetBool(SkipMiddlewareKey) {
		c.Next()
		return
	}

	if c.GetHeader("Authorization") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.New("authorization header is required")})
		c.Writer.Flush()
		c.Abort()
		return
	}
	c.Next()
}

func (s *RestAPI) ValidateToken(c *gin.Context) {

	c.Next()
}
func (api *RestAPI) CORSConfig(c *gin.Context) {
	corsConfig = cors.Config{
		AllowOrigins:  []string{"*"}, // Adicione a origem do seu frontend
		AllowMethods:  []string{"*"},
		AllowHeaders:  []string{"*"},
		ExposeHeaders: []string{"Content-Length"},
		// AllowCredentials: true,
		MaxAge: 12 * time.Hour,
	}
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
	c.Writer.Header().Set("Content-Type", "application/json")

	api.Route.Use(cors.New(corsConfig))
	c.Next()
}

// CORSMiddleware cria um middleware Gin para lidar com CORS.
func (s *RestAPI) CORSMiddleware(c *gin.Context) {
	// Domínios permitidos. Para desenvolvimento, "*" pode ser usado,
	// mas para produção, especifique o domínio do seu frontend.
	// Ex: "https://sua-app-frontend.web.app"
	c.Writer.Header().Set("Access-Control-Allow-Origin", "h*")
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
	c.Writer.Header().Set("Content-Type", "application/json")

	// Métodos HTTP permitidos
	c.Writer.Header().Set("Access-Control-Allow-Methods", "*")

	// Cabeçalhos permitidos
	// Importante: Adicione todos os cabeçalhos personalizados que seu frontend envia.
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

	// Permitir credenciais (se você usar cookies ou autenticação HTTP)
	// c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	// Cache da preflight request (em segundos)
	c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 horas

	// Se for uma requisição OPTIONS (preflight), retorne 204 No Content
	// if c.Request.Method == "OPTIONS" {
	// 	c.AbortWithStatus(http.StatusNoContent) // Ou http.StatusOK se preferir
	// 	return
	// }

	// Processa a próxima requisição
	c.Next()
}
