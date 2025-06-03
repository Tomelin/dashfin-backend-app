package main

import (
	"errors"
	"log"

	"github.com/Tomelin/dashfin-backend-app/config"
	"github.com/Tomelin/dashfin-backend-app/internal/handler/web"
	cryptdata "github.com/Tomelin/dashfin-backend-app/pkg/cryptData"
	"github.com/Tomelin/dashfin-backend-app/pkg/http_server"
	"github.com/go-viper/mapstructure/v2"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	apiResponse, err := loadWebServer(cfg.Fields["webserver"].(map[string]interface{}))
	if err != nil {
		log.Fatal(err)
	}

	dataEncrypt,err := getEncryptToken(cfg.Fields["token"].(string))
	if err != nil {
		log.Fatal(err)
	}

	web.InicializationProfileHandlerHttp("ok", cryptdata.CryptDataInterface,apiResponse.RouterGroup, apiResponse.CorsMiddleware(), apiResponse.MiddlewareHeader)
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


func getEncryptToken(token string) (cryptdata.CryptDataInterface, error) {

	if token == "" {
		return nil,errors.New("token is nil")
	}

	dresult, err :=cryptdata.InicializationCryptData(&token)
	if err != nil {
		return "", err
	}

	

	return dresult, nil
}