package service

import (
	"context"
	"testing"

	entity_profile "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
)

type mockProfileRepository struct{}

func (m *mockProfileRepository) CreateProfile(ctx context.Context, data *entity_profile.Profile) (*entity_profile.Profile, error) {
	data.ID = "test-id"
	return data, nil
}

func (m *mockProfileRepository) GetProfileByID(ctx context.Context, id *string) (*entity_profile.Profile, error) {
	return &entity_profile.Profile{
		ID:             "test-id",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		UserProviderID: *id,
	}, nil
}

func (m *mockProfileRepository) GetProfile(ctx context.Context) ([]entity_profile.Profile, error) {
	return []entity_profile.Profile{
		{
			ID:        "test-id",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		},
	}, nil
}

func (m *mockProfileRepository) GetByFilter(ctx context.Context, data map[string]interface{}) ([]entity_profile.Profile, error) {
	if email, ok := data["Email"].(string); ok && email == "existing@example.com" {
		return []entity_profile.Profile{
			{
				ID:        "existing-id",
				FirstName: "Existing",
				LastName:  "User",
				Email:     "existing@example.com",
			},
		}, nil
	}
	return []entity_profile.Profile{}, nil
}

func (m *mockProfileRepository) UpdateProfile(ctx context.Context, data *entity_profile.Profile) (*entity_profile.Profile, error) {
	return data, nil
}

func BenchmarkProfileServiceCreateProfile(b *testing.B) {
	mockRepo := &mockProfileRepository{}
	service, _ := InicializeProfileService(mockRepo)
	ctx := context.Background()

	profile := &entity_profile.Profile{
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		UserProviderID: "provider-123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.CreateProfile(ctx, profile)
	}
}

func BenchmarkProfileServiceGetProfileByID(b *testing.B) {
	mockRepo := &mockProfileRepository{}
	service, _ := InicializeProfileService(mockRepo)
	ctx := context.Background()
	id := "provider-123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetProfileByID(ctx, &id)
	}
}

func BenchmarkProfileServiceUpdateProfile(b *testing.B) {
	mockRepo := &mockProfileRepository{}
	service, _ := InicializeProfileService(mockRepo)
	ctx := context.Background()

	profile := &entity_profile.Profile{
		ID:             "test-id",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		UserProviderID: "provider-123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.UpdateProfile(ctx, profile)
	}
}

func BenchmarkProfileServiceGetByFilter(b *testing.B) {
	mockRepo := &mockProfileRepository{}
	service, _ := InicializeProfileService(mockRepo)
	ctx := context.Background()

	filter := map[string]interface{}{
		"userProviderID": "provider-123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetByFilter(ctx, filter)
	}
}

func BenchmarkProfileServiceGetProfile(b *testing.B) {
	mockRepo := &mockProfileRepository{}
	service, _ := InicializeProfileService(mockRepo)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetProfile(ctx)
	}
}

func BenchmarkProfileGoalsServiceUpdateProfileGoals(b *testing.B) {
	mockRepo := &mockProfileRepository{}
	profileService, _ := InicializeProfileService(mockRepo)
	goalsService, _ := InicializeProfileGoalsService(mockRepo, profileService)
	ctx := context.Background()
	userId := "provider-123"

	goals := &entity_profile.ProfileGoals{
		Goals2Years: []entity_profile.Goals{
			{
				Name:         "Test Goal",
				TargetDate:   "2025-12-31",
				Description:  "Test Description",
				TargetAmount: 10000.0,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = goalsService.UpdateProfileGoals(ctx, &userId, goals)
	}
}

func BenchmarkProfileProfessionServiceUpdateProfileProfession(b *testing.B) {
	mockRepo := &mockProfileRepository{}
	profileService, _ := InicializeProfileService(mockRepo)
	professionService, _ := InicializeProfileProfessionService(mockRepo, profileService)
	ctx := context.Background()
	userId := "provider-123"

	profession := &entity_profile.ProfileProfession{
		Profession:    "Software Engineer",
		Company:       "Tech Corp",
		ContractType:  entity_profile.ContractTypeCTL,
		MonthlyIncome: 5000.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = professionService.UpdateProfileProfession(ctx, &userId, profession)
	}
}
