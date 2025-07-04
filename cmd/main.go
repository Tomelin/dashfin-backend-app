package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Tomelin/dashfin-backend-app/config"
	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	entity_platform "github.com/Tomelin/dashfin-backend-app/internal/core/entity/platform"
	"github.com/Tomelin/dashfin-backend-app/internal/core/repository"
	repository_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/repository/dashboard"
	repository_finance "github.com/Tomelin/dashfin-backend-app/internal/core/repository/finance"
	repository_platform "github.com/Tomelin/dashfin-backend-app/internal/core/repository/platform"
	repository_profile "github.com/Tomelin/dashfin-backend-app/internal/core/repository/profile"
	"github.com/Tomelin/dashfin-backend-app/internal/core/service"
	service_dashboard "github.com/Tomelin/dashfin-backend-app/internal/core/service/dashboard"
	service_finance "github.com/Tomelin/dashfin-backend-app/internal/core/service/finance"
	service_platform "github.com/Tomelin/dashfin-backend-app/internal/core/service/platform"
	service_profile "github.com/Tomelin/dashfin-backend-app/internal/core/service/profile"
	"github.com/Tomelin/dashfin-backend-app/internal/handler/web"
	web_dashboard "github.com/Tomelin/dashfin-backend-app/internal/handler/web/dashboard"
	web_finance "github.com/Tomelin/dashfin-backend-app/internal/handler/web/finance"
	web_finance_expense "github.com/Tomelin/dashfin-backend-app/internal/handler/web/finance/expense"
	web_finance_income "github.com/Tomelin/dashfin-backend-app/internal/handler/web/finance/income"
	web_report "github.com/Tomelin/dashfin-backend-app/internal/handler/web/finance/report"
	web_platform "github.com/Tomelin/dashfin-backend-app/internal/handler/web/platform"
	"github.com/Tomelin/dashfin-backend-app/pkg/authenticatior"
	"github.com/Tomelin/dashfin-backend-app/pkg/cache" // Added cache import
	cryptdata "github.com/Tomelin/dashfin-backend-app/pkg/cryptData"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
	"github.com/Tomelin/dashfin-backend-app/pkg/http_server" // Added Redis import
	"github.com/Tomelin/dashfin-backend-app/pkg/message_queue"
	"github.com/go-viper/mapstructure/v2"
	// Ensure this is imported for adapters
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
	log.Println("starting firebase", cfg.Fields["firebase"])
	if err != nil {
		log.Fatal(err)
	}

	// Import data at firestore
	// iif := database.NewFirebaseInsert(db)
	// err = iif.InsertBrazilianBanksFromJSON(context.Background())
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Fatalln("finish")

	cacheClient, err := initializeCache(cfg.Fields["cache"])
	if err != nil {
		log.Fatal(err)
	}

	mq, err := initializeMessageQueue(cfg.Fields["message_queue"].(map[string]interface{}))
	if err != nil {
		log.Fatal(err)
	}
	defer mq.Close()

	svcProfilePerson, svcProfileProfession, svcProfileGoals, err := initializeProfileServices(db)
	if err != nil {
		log.Fatal(err)
	}

	// Reconstruct the aggregate ProfileAllService for handlers that need it
	svcProfileAll, err := service_profile.InicializeProfileAllService(svcProfilePerson, svcProfileProfession, svcProfileGoals)
	if err != nil {
		log.Fatalf("failed to initialize ProfileAllService: %v", err)
	}

	svcSupport, err := initializeSupportServices(db)
	if err != nil {
		log.Fatal(err)
	}

	svcFinancialInstitution, err := initializeFinancialInstitution(db)
	if err != nil {
		log.Fatal(err)
	}

	svcExpenseRecord, err := initializeExpenseRecordServices(db, mq)
	if err != nil {
		log.Fatal(err)
	}

	svcBankAccount, err := initializeBankAccountServices(db)
	if err != nil {
		log.Fatal(err)
	}

	svcCreditCard, err := initializeCreditCardServices(db)
	if err != nil {
		log.Fatal(err)
	}

	svcIncomeRecord, err := initializeIncomeRecordServices(db, mq)
	if err != nil {
		log.Fatal(err)
	}

	svcSpendingRecord, err := initializeSpendingPlanServices(db, cacheClient)
	if err != nil {
		log.Fatal(err)
	}

	srvDashboard, err := initializeDashboardServices(svcBankAccount, svcExpenseRecord, svcIncomeRecord, svcProfileGoals, svcFinancialInstitution, mq, db)
	if err != nil {
		log.Fatal(err)
	}

	svcReport, err := initializeReportServices(svcIncomeRecord, svcExpenseRecord, cacheClient, mq)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Financial Repository & Service for PlannedVsActual
	financialRepo, _ := repository_dashboard.NewFirebaseFinancialRepository(db)
	financialSvc, _ := service_dashboard.NewFinancialService(financialRepo, cacheClient, svcExpenseRecord, svcSpendingRecord)
	// No error is returned by NewFinancialService or NewFirebaseFinancialRepository, so no error check needed here.

	web_dashboard.InitializeDashboardHandler(srvDashboard, crypt, authClient, apiResponse.RouterGroup, apiResponse.CorsMiddleware(), apiResponse.MiddlewareHeader)
	web_dashboard.InitializePlannedVsActualHandler(financialSvc, authClient, crypt, apiResponse.RouterGroup, apiResponse.CorsMiddleware(), apiResponse.MiddlewareHeader)
	web.InicializationProfileHandlerHttp(svcProfileAll, crypt, authClient, apiResponse.RouterGroup, apiResponse.CorsMiddleware(), apiResponse.MiddlewareHeader)
	web.InicializationSupportHandlerHttp(svcSupport, crypt, authClient, apiResponse.RouterGroup, apiResponse.CorsMiddleware(), apiResponse.MiddlewareHeader)
	web_platform.NewFinancialInstitutionHandler(svcFinancialInstitution, crypt, authClient, apiResponse.RouterGroup, apiResponse.CorsMiddleware(), apiResponse.MiddlewareHeader)
	web_finance_expense.InitializeExpenseRecordHandler(svcExpenseRecord, crypt, authClient, apiResponse.RouterGroup, apiResponse.CorsMiddleware(), apiResponse.MiddlewareHeader)
	web_finance.InitializeBankAccountHandler(svcBankAccount, crypt, authClient, apiResponse.RouterGroup, apiResponse.CorsMiddleware(), apiResponse.MiddlewareHeader)
	web_finance.InitializeCreditCardHandler(svcCreditCard, crypt, authClient, apiResponse.RouterGroup, apiResponse.CorsMiddleware(), apiResponse.MiddlewareHeader)
	web_finance_income.InitializeIncomeRecordHandler(svcIncomeRecord, crypt, authClient, apiResponse.RouterGroup, apiResponse.CorsMiddleware(), apiResponse.MiddlewareHeader)
	web_finance.InitializeSpendingPlanHandler(svcSpendingRecord, crypt, authClient, apiResponse.RouterGroup, apiResponse.CorsMiddleware(), apiResponse.MiddlewareHeader)
	web_report.InitializeReportHandler(svcReport, crypt, authClient, apiResponse.RouterGroup, apiResponse.CorsMiddleware(), apiResponse.MiddlewareHeader)

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

func initializeCache(fields interface{}) (cache.CacheService, error) {

	b, _ := json.Marshal(fields)

	var config cache.CacheConfig
	err := json.Unmarshal(b, &config)
	if err != nil {
		return nil, err
	}

	cacheClient, err := cache.NewRedisCacheService(config)
	if err != nil {
		return nil, err
	}

	return cacheClient, nil

}

func initializeMessageQueue(fields map[string]interface{}) (message_queue.MessageQueue, error) {

	b, _ := json.Marshal(fields)

	var config message_queue.Config
	err := json.Unmarshal(b, &config)
	if err != nil {
		return nil, err
	}

	mq, err := message_queue.NewRabbitMQ(config)
	if err != nil {
		return nil, err
	}

	err = mq.Setup()
	if err != nil {
		return nil, err
	}

	return mq, nil
}

func initializeCryptData(encryptField interface{}) (cryptdata.CryptDataInterface, error) {
	token := fmt.Sprintf("%v", encryptField)
	return cryptdata.InicializationCryptData(&token)
}

func initializeFirebase(firebaseField interface{}) (authenticatior.Authenticator, database.FirebaseDBInterface, error) {
	var fConfig authenticatior.FirebaseConfig
	log.Println("firebaseConfig", firebaseField)
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

func initializeProfileServices(db database.FirebaseDBInterface) (service_profile.ProfilePersonServiceInterface, service_profile.ProfileProfessionServiceInterface, service_profile.ProfileGoalsServiceInterface, error) {
	repoProfile, err := repository_profile.InicializeProfileRepository(db)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize profile repository: %w", err)
	}

	svcProfilePerson, err := service_profile.InicializeProfileService(repoProfile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize profile person service: %w", err)
	}

	svcProfileProfession, err := service_profile.InicializeProfileProfessionService(repoProfile, svcProfilePerson)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize profile profession service: %w", err)
	}

	svcProfileGoals, err := service_profile.InicializeProfileGoalsService(repoProfile, svcProfilePerson)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize profile goals service: %w", err)
	}

	return svcProfilePerson, svcProfileProfession, svcProfileGoals, nil
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

func initializeExpenseRecordServices(db database.FirebaseDBInterface, mq message_queue.MessageQueue) (entity_finance.ExpenseRecordServiceInterface, error) {
	repoExpenseRecord, err := repository_finance.InitializeExpenseRecordRepository(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize expense record repository: %w", err)
	}

	svcExpenseRecord, err := service_finance.InitializeExpenseRecordService(repoExpenseRecord, mq)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize expense record service: %w", err)
	}
	return svcExpenseRecord, nil
}

func initializeBankAccountServices(db database.FirebaseDBInterface) (entity_finance.BankAccountServiceInterface, error) {
	repoBankAccount, err := repository_finance.InitializeBankAccountRepository(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize bank account repository: %w", err)
	}
	return service_finance.InitializeBankAccountService(repoBankAccount)
}

func initializeCreditCardServices(db database.FirebaseDBInterface) (entity_finance.CreditCardServiceInterface, error) {
	repoCreditCard, err := repository_finance.InitializeCreditCardRepository(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize credit card repository: %w", err)
	}
	return service_finance.InitializeCreditCardService(repoCreditCard)

}

func initializeIncomeRecordServices(db database.FirebaseDBInterface, mq message_queue.MessageQueue) (entity_finance.IncomeRecordServiceInterface, error) {
	repoIncomeRecord, err := repository_finance.InitializeIncomeRecordRepository(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize income record repository: %w", err)
	}

	svcIncomeRecord, err := service_finance.InitializeIncomeRecordService(repoIncomeRecord, mq)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize income record service: %w", err)
	}
	return svcIncomeRecord, nil
}

func initializeSpendingPlanServices(db database.FirebaseDBInterface, cache cache.CacheService) (entity_finance.SpendingPlanServiceInterface, error) {
	repoSpendingRecord, err := repository_finance.InitializeSpendingPlanRepository(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize income record repository: %w", err)
	}

	svcSpendingRecord, err := service_finance.InitializeSpendingPlanService(repoSpendingRecord, cache)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize income record service: %w", err)
	}
	return svcSpendingRecord, nil
}

func initializeReportServices(
	income entity_finance.IncomeRecordServiceInterface,
	expense entity_finance.ExpenseRecordServiceInterface,
	cache cache.CacheService,
	messageQueue message_queue.MessageQueue,
) (entity_finance.FinancialReportDataServiceInterface, error) {

	svcReport, err := service_finance.InitializeFinancialReportDataService(income, expense, cache, messageQueue)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize report service: %w", err)
	}
	return svcReport, nil
}

func initializeDashboardServices(
	bankAccountSvc entity_finance.BankAccountServiceInterface,
	expenseRecordSvc entity_finance.ExpenseRecordServiceInterface,
	incomeRecordSvc entity_finance.IncomeRecordServiceInterface,
	profileGoalsSvc service_profile.ProfileGoalsServiceInterface,
	platformInst entity_platform.FinancialInstitutionInterface,
	messageQueue message_queue.MessageQueue,
	db database.FirebaseDBInterface,
) (*service_dashboard.DashboardService, error) {
	repoSpendingRecord := repository_dashboard.NewInMemoryDashboardRepository(db)

	svcSpendingRecord := service_dashboard.NewDashboardService(
		bankAccountSvc,
		expenseRecordSvc,
		incomeRecordSvc,
		profileGoalsSvc,
		repoSpendingRecord,
		messageQueue,
		platformInst,
	)

	return svcSpendingRecord, nil
}
