package entity_profile

import (
	"testing"
	"time"
)

func BenchmarkProfileValidate(b *testing.B) {
	profile := &Profile{
		ID:             "test-id",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john.doe@example.com",
		Phone:          "+1234567890",
		BirthDate:      "1990-01-01",
		Sexo:           "M",
		Cep:            "12345-678",
		City:           "Test City",
		State:          "TS",
		UserProviderID: "provider-123",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = profile.Validate()
	}
}

func BenchmarkProfileValidateBirthDate(b *testing.B) {
	profile := &Profile{}
	dateString := "1990-01-01"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = profile.ValidateBirthDate(dateString)
	}
}

func BenchmarkProfileSetBirthDate(b *testing.B) {
	profile := &Profile{}
	dateString := "1990-01-01"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = profile.SetBirthDate(dateString)
	}
}

func BenchmarkProfileStructCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profile := Profile{
			ID:             "test-id",
			FirstName:      "John",
			LastName:       "Doe",
			Email:          "john.doe@example.com",
			Phone:          "+1234567890",
			BirthDate:      "1990-01-01",
			Sexo:           "M",
			Cep:            "12345-678",
			City:           "Test City",
			State:          "TS",
			UserProviderID: "provider-123",
			Profession: ProfileProfession{
				Profession:    "Software Engineer",
				Company:       "Tech Corp",
				ContractType:  ContractTypeCTL,
				MonthlyIncome: 5000.0,
			},
			Goals: ProfileGoals{
				Goals2Years: []Goals{
					{
						Name:         "Buy a car",
						TargetDate:   "2025-12-31",
						Description:  "Purchase a new vehicle",
						TargetAmount: 30000.0,
					},
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = profile
	}
}
