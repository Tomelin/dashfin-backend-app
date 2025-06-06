package http_server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type Certificate struct {
	Key string `mapstructure:"key"`
	Crt string `mapstructure:"crt"`
}

type RestAPIConfig struct {
	Port            string `mapstructure:"port"`
	SSLEnabled      bool   `mapstructure:"ssl_enabled"`
	Host            string `mapstructure:"host"`
	Version         string `mapstructure:"version"`
	Name            string `mapstructure:"name"`
	CertificateCrt  string `mapstructure:"certificate_crt"`
	CertificateKey  string `mapstructure:"certificate_key"`
	Token           string `mapstructure:"token"`
	OtelServiceName string `mapstructure:"otel_service_name"`
	Mode            string `mapstructure:"mode"`
}

type RestAPI struct {
	Config *RestAPIConfig
	Route  *gin.Engine
	*gin.RouterGroup
}

func NewRestApi(fields RestAPIConfig) (*RestAPI, error) {

	b, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}

	var rest *RestAPIConfig
	err = json.Unmarshal(b, &rest)
	if err != nil {
		return nil, err
	}

	r, g := newRestAPI(rest)

	return &RestAPI{
		Config:      rest,
		Route:       r,
		RouterGroup: g,
	}, nil
}

func (api *RestAPIConfig) Validate() error {

	if api.Port == "" {
		api.Port = "8080"
	}
	if api.Host == "" {
		api.Host = "0.0.0.0"
	}
	if api.Name == "" {
		api.Name = "api"
	}
	if api.Version == "" {
		api.Version = "v1"
	}
	if api.Mode == "" {
		api.Mode = "debug"
	}
	if api.SSLEnabled {
		if api.CertificateCrt == "" {
			return fmt.Errorf("certificate_crt is required")
		}
		if api.CertificateKey == "" {
			return fmt.Errorf("certificate_key is required")
		}
	}

	return nil
}

var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "cloud-collector-resources3",
	Description:      "",
	InfoInstanceName: "swagger",
	// SwaggerTemplate:  docTemplate,
	// LeftDelim:  "{{",
	// RightDelim: "}}",
}

func newRestAPI(config *RestAPIConfig) (*gin.Engine, *gin.RouterGroup) {

	if config == nil {
		config = &RestAPIConfig{}
	}
	serviceName := "app_name"
	if config.OtelServiceName != "" {
		serviceName = config.OtelServiceName
	}

	router := gin.New()
	if config.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router.Use(otelgin.Middleware(serviceName))
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// if sesseionStore != nil {
	// 	router.Use(sessions.Sessions("Authorization", *sesseionStore))
	// }

	router.UseH2C = true

	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"127.0.0.1", "192.168.1.2", "10.0.0.0/8"})

	routerGroupPath := fmt.Sprintf("/%s", config.Name)
	routerPath = router.Group(routerGroupPath)

	router.GET("/metrics", prometheusHandler())

	// Set swagger
	routerPath.GET("/docs/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.DefaultModelsExpandDepth(-1)))
	routerPath.GET("/docs/swagger", func(c *gin.Context) {
		c.Redirect(301, fmt.Sprintf("%s/docs/swagger/index.html", routerGroupPath))
	})
	routerPath.GET("/docs", func(c *gin.Context) {

		c.Redirect(301, fmt.Sprintf("%s/docs/swagger/index.html", routerGroupPath))
	})
	routerPath.GET("/", func(c *gin.Context) {
		c.Redirect(301, fmt.Sprintf("%s/docs/swagger/index.html", routerGroupPath))
	})

	router.Use(setHeader)

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", " X-Authorization", " X-USERID", " X-APP", " X-USER", " X-TRACE-ID"}
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", " X-Authorization", " X-USERID", " X-APP", " X-USER", " X-TRACE-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Use(cors.New(corsConfig))

	return router, routerPath

}
