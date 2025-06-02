package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Tomelin/dashfin-backend-app/config"
	"github.com/Tomelin/dashfin-backend-app/internal/handler/web"
	"github.com/Tomelin/dashfin-backend-app/pkg/http_server"
	"github.com/go-viper/mapstructure/v2"
)

func main() {

	demo()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	apiResponse, err := loadWebServer(cfg.Fields["webserver"].(map[string]interface{}))
	if err != nil {
		log.Fatal(err)
	}

	web.InicializationProfileHandlerHttp("ok", apiResponse.RouterGroup)
	err = apiResponse.Run(apiResponse.Route.Handler())
	if err != nil {
		log.Fatal(err)
	}
}

func loadWebServer(fields map[string]interface{}) (*http_server.RestAPI, error) {

	var apiConfig http_server.RestAPIConfig
	err := mapstructure.Decode(fields, &apiConfig)
	if err != nil {
		return nil, err
	}

	log.Println(apiConfig.Validate())

	api, err := http_server.NewRestApi(apiConfig)
	if err != nil {
		return nil, err
	}
	return api, nil

}

func demo() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello, world")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Handling HTTP requests on %s.", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))

}
