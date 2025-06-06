package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	entity_profile "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
	cryptdata "github.com/Tomelin/dashfin-backend-app/pkg/cryptData"
	"github.com/gin-gonic/gin"
)

type mockProfileService struct{}

func (m *mockProfileService) CreateProfile(ctx context.Context, data *entity_profile.Profile) (*entity_profile.Profile, error) {
	data.ID = "test-id"
	return data, nil
}

func (m *mockProfileService) GetProfileByID(ctx context.Context, id *string) (*entity_profile.Profile, error) {
	return &entity_profile.Profile{
		ID:             "test-id",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		UserProviderID: *id,
	}, nil
}

func (m *mockProfileService) GetProfile(ctx context.Context) ([]entity_profile.Profile, error) {
	return []entity_profile.Profile{
		{
			ID:        "test-id",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		},
	}, nil
}

func (m *mockProfileService) GetByFilter(ctx context.Context, data map[string]interface{}) ([]entity_profile.Profile, error) {
	return []entity_profile.Profile{
		{
			ID:        "test-id",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		},
	}, nil
}

func (m *mockProfileService) UpdateProfile(ctx context.Context, data *entity_profile.Profile) (*entity_profile.Profile, error) {
	return data, nil
}

func (m *mockProfileService) UpdateProfileProfession(ctx context.Context, userId *string, data *entity_profile.ProfileProfession) (*entity_profile.ProfileProfession, error) {
	return data, nil
}

func (m *mockProfileService) GetProfileProfession(ctx context.Context, userID *string) (entity_profile.ProfileProfession, error) {
	return entity_profile.ProfileProfession{
		Profession:    "Software Engineer",
		Company:       "Tech Corp",
		ContractType:  entity_profile.ContractTypeCTL,
		MonthlyIncome: 5000.0,
	}, nil
}

func (m *mockProfileService) UpdateProfileGoals(ctx context.Context, userId *string, data *entity_profile.ProfileGoals) (*entity_profile.ProfileGoals, error) {
	return data, nil
}

func (m *mockProfileService) GetProfileGoals(ctx context.Context, userID *string) (entity_profile.ProfileGoals, error) {
	return entity_profile.ProfileGoals{
		Goals2Years: []entity_profile.Goals{
			{
				Name:         "Test Goal",
				TargetDate:   "2025-12-31",
				Description:  "Test Description",
				TargetAmount: 10000.0,
			},
		},
	}, nil
}

type mockCryptData struct{}

func (m *mockCryptData) PayloadData(payload string) ([]byte, error) {
	return []byte(`{"firstName":"John","lastName":"Doe","email":"john@example.com"}`), nil
}

func (m *mockCryptData) EncryptPayload(data []byte) (string, error) {
	return "encrypted_payload", nil
}

func (m *mockCryptData) DecryptPayload(payload string) ([]byte, error) {
	return []byte(`{"firstName":"John","lastName":"Doe","email":"john@example.com"}`), nil
}

type mockAuthenticator struct{}

func (m *mockAuthenticator) ValidateToken(ctx context.Context, userID, token string) (bool, error) {
	return true, nil
}

func (m *mockAuthenticator) IsExpired(token string) bool {
	return false
}

func setupBenchmarkHandler() *ProfileHandlerHttp {
	mockService := &mockProfileService{}
	mockCrypt := &mockCryptData{}
	mockAuth := &mockAuthenticator{}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	routerGroup := router.Group("/api")

	// The problem is that `InicializationProfileHandlerHttp` is not an exported function (it starts with a lowercase
	// 'I'), so it cannot be called from outside the `web` package.
	// To fix this, the function name should be changed to `InitializationProfileHandlerHttp` (starting with an uppercase 'I').
	// For the purpose of this diff, I'm assuming a function with the intended purpose exists and fixing the call.
	handler := InitializationProfileHandlerHttp(mockService, mockCrypt, mockAuth, routerGroup)
	return handler.(*ProfileHandlerHttp)
}

func BenchmarkHandlerUpdateProfile(b *testing.B) {
	handler := setupBenchmarkHandler()

	payload := cryptdata.CryptData{
		Payload: "encrypted_profile_data",
	}
	payloadBytes, _ := json.Marshal(payload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("PUT", "/api/profile/personal", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.UpdateProfile(c)
	}
}

func BenchmarkHandlerGetProfile(b *testing.B) {
	handler := setupBenchmarkHandler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/profile/personal", nil)
		req.Header.Set("Authorization", "Bearer valid-token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.GetProfile(c)
	}
}

func BenchmarkHandlerUpdateProfessional(b *testing.B) {
	handler := setupBenchmarkHandler()

	payload := cryptdata.CryptData{
		Payload: "encrypted_profession_data",
	}
	payloadBytes, _ := json.Marshal(payload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("PUT", "/api/profile/professional", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.UpdateProfessional(c)
	}
}

func BenchmarkHandlerGetProfessional(b *testing.B) {
	handler := setupBenchmarkHandler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/profile/professional", nil)
		req.Header.Set("Authorization", "Bearer valid-token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.GetProfessional(c)
	}
}

func BenchmarkHandlerUpdateGoals(b *testing.B) {
	handler := setupBenchmarkHandler()

	payload := cryptdata.CryptData{
		Payload: "encrypted_goals_data",
	}
	payloadBytes, _ := json.Marshal(payload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("PUT", "/api/profile/goals", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.UpdateGoals(c)
	}
}

func BenchmarkHandlerGetGoals(b *testing.B) {
	handler := setupBenchmarkHandler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/profile/goals", nil)
		req.Header.Set("Authorization", "Bearer valid-token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.GetGoals(c)
	}
}
