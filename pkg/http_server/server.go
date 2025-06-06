package http_server

import (
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

	srv := http.Server{
		Addr:    ":" + s.Config.Port,
		Handler: s.Route.Handler(),
	}

	s.Route.Use(s.CorsMiddleware())
	http2.ConfigureServer(&srv, &http2.Server{})
	s.Route.Use(s.EmbeddedMiddleware)

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

func (s *RestAPI) EmbeddedMiddleware(c *gin.Context) {

	if c.Request.Method == "OPTIONS" {
		c.Status(http.StatusNoContent)
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	if c.GetHeader("X-APP") != "DashfinApp" {
		c.AbortWithStatus(401)
		return
	}

	c.Next()
}

func (s *RestAPI) MiddlewareHeader(c *gin.Context) {

	if c.GetHeader("X-USERID") == "" {
		c.AbortWithStatus(401)
		return
	}
	if c.GetHeader("X-AUTHORIZATION") == "" {
		c.AbortWithStatus(401)
		return
	}

	c.Next()
}
func (s *RestAPI) CorsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
