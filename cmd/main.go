package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Tomelin/dashfin-backend-app/config"
	entity_platform "github.com/Tomelin/dashfin-backend-app/internal/core/entity/platform"
	"github.com/Tomelin/dashfin-backend-app/internal/core/repository"
	repository_platform "github.com/Tomelin/dashfin-backend-app/internal/core/repository/platform"
	repository_profile "github.com/Tomelin/dashfin-backend-app/internal/core/repository/profile"
	"github.com/Tomelin/dashfin-backend-app/internal/core/service"
	service_platform "github.com/Tomelin/dashfin-backend-app/internal/core/service/platform"
	service_profile "github.com/Tomelin/dashfin-backend-app/internal/core/service/profile"
	"github.com/Tomelin/dashfin-backend-app/internal/handler/web"
	web_platform "github.com/Tomelin/dashfin-backend-app/internal/handler/web/platform"
	"github.com/Tomelin/dashfin-backend-app/pkg/authenticatior"
	cryptdata "github.com/Tomelin/dashfin-backend-app/pkg/cryptData"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
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

	crypt, err := initializeCryptData(cfg.Fields["encrypt"])
	if err != nil {
		log.Fatal(err)
	}

	authClient, db, err := initializeFirebase(cfg.Fields["firebase"])
	if err != nil {
		log.Fatal(err)
	}

	// Import data at firestore
//	iif := database.NewFirebaseInsert(db)
//	err = iif.InsertBrazilianBanksFromJSON(context.Background())
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Fatal("finish")

	svcProfile, err := initializeProfileServices(db)
	if err != nil {
		log.Fatal(err)
	}

	svcSupport, err := initializeSupportServices(db)
	if err != nil {
		log.Fatal(err)
	}

	svcFinancialInstitution, err := initializeFinancialInstitution(db)
	if err != nil {
		log.Fatal(err)
	}

	web.InicializationProfileHandlerHttp(svcProfile, crypt, authClient, apiResponse.RouterGroup, apiResponse.CorsMiddleware(), apiResponse.MiddlewareHeader)
	web.InicializationSupportHandlerHttp(svcSupport, crypt, authClient, apiResponse.RouterGroup, apiResponse.CorsMiddleware(), apiResponse.MiddlewareHeader)
	web_platform.NewFinancialInstitutionHandler(svcFinancialInstitution, crypt, authClient, apiResponse.RouterGroup, apiResponse.CorsMiddleware(), apiResponse.MiddlewareHeader)

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

func initializeCryptData(encryptField interface{}) (cryptdata.CryptDataInterface, error) {
	token := fmt.Sprintf("%v", encryptField)
	return cryptdata.InicializationCryptData(&token)
}

func initializeFirebase(firebaseField interface{}) (authenticatior.Authenticator, database.FirebaseDBInterface, error) {
	var fConfig authenticatior.FirebaseConfig
	b, err := json.Marshal(firebaseField)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal firebase config: %w", err)
	}

	if err := json.Unmarshal(b, &fConfig); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal firebase config: %w", err)
	}

	authClient, err := authenticatior.InitializeAuth(context.Background(), &fConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize auth: %w", err)
	}

	db, err := database.InitializeFirebaseDB(database.FirebaseConfig{
		ProjectID:             fConfig.ProjectID,
		APIKey:                fConfig.APIKey,
		DatabaseURL:           fConfig.DatabaseURL,
		StorageBucket:         fConfig.StorageBucket,
		AppID:                 fConfig.AppID,
		AuthDomain:            fConfig.AuthDomain,
		MessagingSenderID:     fConfig.MessagingSenderID,
		ServiceAccountKeyPath: fConfig.ServiceAccountKeyPath,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize firebase DB: %w", err)
	}

	return authClient, db, nil
}

func initializeProfileServices(db database.FirebaseDBInterface) (service_profile.ProfileServiceInterface, error) {
	repoProfile, err := repository_profile.InicializeProfileRepository(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize profile repository: %w", err)
	}

	svcProfilePerson, err := service_profile.InicializeProfileService(repoProfile)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize profile person service: %w", err)
	}

	svcProfileProfession, err := service_profile.InicializeProfileProfessionService(repoProfile, svcProfilePerson)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize profile profession service: %w", err)
	}

	svcProfileGoals, err := service_profile.InicializeProfileGoalsService(repoProfile, svcProfilePerson)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize profile goals service: %w", err)
	}

	return service_profile.InicializeProfileAllService(svcProfilePerson, svcProfileProfession, svcProfileGoals)
}

func initializeSupportServices(db database.FirebaseDBInterface) (service.SupportServiceInterface, error) {
	repoSupport, err := repository.InicializeSupportRepository(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize support repository: %w", err)
	}

	return service.InicializeSupportService(repoSupport)
}

func initializeFinancialInstitution(db database.FirebaseDBInterface) (entity_platform.FinancialInstitutionInterface, error) {
	repoSupport, err := repository_platform.NewFinancialInstitutionRepository(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize support repository: %w", err)
	}

	return service_platform.NewFinancialInstitutionService(repoSupport)
}
