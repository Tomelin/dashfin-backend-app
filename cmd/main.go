package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/Tomelin/dashfin-backend-app/config"
	"github.com/Tomelin/dashfin-backend-app/internal/handler/web"
	"github.com/Tomelin/dashfin-backend-app/pkg/authenticatior"
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

	token := "VGhpc0lzQTE2Qnl0ZUtleVRoaXNJc0ExNkJ5dGVJVgo="
	crypt, err := cryptdata.InicializationCryptData(&token)
	if err != nil {
		log.Fatal(err)
	}

	var fConfig authenticatior.FirebaseConfig
	json.Unmarshal([]byte(cfg.Fields["firebase"].(string)), &fConfig)

	log.Println(fConfig)

	authClient, err := authenticatior.InitializeAuth(context.Background(), &authenticatior.FirebaseConfig{
		ProjectID:         fConfig.ProjectID,
		APIKey:            fConfig.APIKey,
		AuthDomain:        fConfig.AuthDomain,
		AppID:             fConfig.AppID,
		MessagingSenderID: fConfig.MessagingSenderID,
		StorageBucket:     fConfig.StorageBucket,
	})
	log.Println(authClient, err)

	web.InicializationProfileHandlerHttp("ok", crypt, apiResponse.RouterGroup, apiResponse.CorsMiddleware(), apiResponse.MiddlewareHeader)
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

func getEncryptToken(data interface{}) (cryptdata.CryptDataInterface, error) {

	if data == nil {
		return nil, errors.New("token is nil")
	}

	token, ok := data.(string)
	if !ok {
		return nil, errors.New("token is nil")
	}

	dresult, err := cryptdata.InicializationCryptData(&token)
	if err != nil {
		return nil, err
	}

	return dresult, nil
}
