package repository

import (
	"context"
	"testing"

	entity_profile "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
)

type mockFirebaseDB struct{}

func (m *mockFirebaseDB) Create(ctx context.Context, data map[string]interface{}, collection string) ([]byte, error) {
	return []byte(`{"id":"test-id"}`), nil
}

func (m *mockFirebaseDB) Update(ctx context.Context, id string, data map[string]interface{}, collection string) error {
	return nil
}

func (m *mockFirebaseDB) Get(ctx context.Context, collection string) ([]byte, error) {
	return []byte(`[{"id":"test-id","firstName":"John","lastName":"Doe","email":"john@example.com"}]`), nil
}

func (m *mockFirebaseDB) GetByFilter(ctx context.Context, filter map[string]interface{}, collection string) ([]byte, error) {
	return []byte(`[{"id":"test-id","firstName":"John","lastName":"Doe","email":"john@example.com","userProviderID":"provider-123"}]`), nil
}

func (m *mockFirebaseDB) Delete(ctx context.Context, id string, collection string) error {
	return nil
}

func BenchmarkProfileRepositoryCreateProfile(b *testing.B) {
	mockDB := &mockFirebaseDB{}
	repo, _ := InicializeProfileRepository(mockDB)
	ctx := context.Background()
	
	profile := &entity_profile.Profile{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		UserProviderID: "provider-123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.CreateProfile(ctx, profile)
	}
}

func BenchmarkProfileRepositoryGetProfileByID(b *testing.B) {
	mockDB := &mockFirebaseDB{}
	repo, _ := InicializeProfileRepository(mockDB)
	ctx := context.Background()
	id := "provider-123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetProfileByID(ctx, &id)
	}
}

func BenchmarkProfileRepositoryUpdateProfile(b *testing.B) {
	mockDB := &mockFirebaseDB{}
	repo, _ := InicializeProfileRepository(mockDB)
	ctx := context.Background()
	
	profile := &entity_profile.Profile{
		ID:        "test-id",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		UserProviderID: "provider-123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.UpdateProfile(ctx, profile)
	}
}

func BenchmarkProfileRepositoryGetByFilter(b *testing.B) {
	mockDB := &mockFirebaseDB{}
	repo, _ := InicializeProfileRepository(mockDB)
	ctx := context.Background()
	
	filter := map[string]interface{}{
		"userProviderID": "provider-123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByFilter(ctx, filter)
	}
}

func BenchmarkProfileRepositoryGetProfile(b *testing.B) {
	mockDB := &mockFirebaseDB{}
	repo, _ := InicializeProfileRepository(mockDB)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetProfile(ctx)
	}
}

func BenchmarkProfileGoalsRepositoryUpdateProfileGoals(b *testing.B) {
	mockDB := &mockFirebaseDB{}
	repo, _ := InicializeProfileGoalsRepository(mockDB)
	ctx := context.Background()
	
	profile := &entity_profile.Profile{
		ID:        "test-id",
		UserProviderID: "provider-123",
		Goals: entity_profile.ProfileGoals{
			Goals2Years: []entity_profile.Goals{
				{
					Name:         "Test Goal",
					TargetDate:   "2025-12-31",
					Description:  "Test Description",
					TargetAmount: 10000.0,
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.UpdateProfileGoals(ctx, profile)
	}
}

func BenchmarkProfileProfessionRepositoryUpdateProfileProfession(b *testing.B) {
	mockDB := &mockFirebaseDB{}
	repo, _ := InicializeProfileProfessionRepository(mockDB)
	ctx := context.Background()
	
	profile := &entity_profile.Profile{
		ID:        "test-id",
		UserProviderID: "provider-123",
		Profession: entity_profile.ProfileProfession{
			Profession:    "Software Engineer",
			Company:       "Tech Corp",
			ContractType:  entity_profile.ContractTypeCTL,
			MonthlyIncome: 5000.0,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.UpdateProfileProfession(ctx, profile)
	}
}