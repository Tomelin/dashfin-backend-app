package entity_profile

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ProfileTestSuite struct {
	suite.Suite
}

func TestProfileSuite(t *testing.T) {
	suite.Run(t, new(ProfileTestSuite))
}

func (suite *ProfileTestSuite) TestContractTypeValues() {
	tests := []struct {
		name     string
		value    ContractTypeValues
		expected string
	}{
		{"CTL Contract", ContractTypeCTL, "clt"},
		{"PJ Contract", ContractTypePJ, "pj"},
		{"Temporary Contract", ContractTypeTemporary, "temporary"},
		{"Internship Contract", ContractTypeInternship, "internship"},
		{"Public Servant Contract", ContractTypePublicServant, "public_servant"},
		{"Other Contract", ContractTypeOther, "other"},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			assert.Equal(suite.T(), tt.expected, string(tt.value))
		})
	}
}

func (suite *ProfileTestSuite) TestProfileProfession() {
	tests := []struct {
		name       string
		profession ProfileProfession
		expected   ProfileProfession
	}{
		{
			name: "Valid ProfileProfession",
			profession: ProfileProfession{
				Profession:    "Software Engineer",
				Company:       "TechCorp",
				ContractType:  ContractTypeCTL,
				MonthlyIncome: 5000.0,
			},
			expected: ProfileProfession{
				Profession:    "Software Engineer",
				Company:       "TechCorp",
				ContractType:  ContractTypeCTL,
				MonthlyIncome: 5000.0,
			},
		},
		{
			name: "ProfileProfession with minimal data",
			profession: ProfileProfession{
				Profession:   "Developer",
				ContractType: ContractTypePJ,
			},
			expected: ProfileProfession{
				Profession:   "Developer",
				ContractType: ContractTypePJ,
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			assert.Equal(suite.T(), tt.expected, tt.profession)
		})
	}
}

func (suite *ProfileTestSuite) TestGoals() {
	tests := []struct {
		name     string
		goal     Goals
		expected Goals
	}{
		{
			name: "Complete Goals",
			goal: Goals{
				Name:         "Buy a house",
				TargetDate:   "2025-12-31",
				Description:  "Purchase first home",
				TargetAmount: 300000.0,
			},
			expected: Goals{
				Name:         "Buy a house",
				TargetDate:   "2025-12-31",
				Description:  "Purchase first home",
				TargetAmount: 300000.0,
			},
		},
		{
			name: "Minimal Goals",
			goal: Goals{
				Name: "Emergency Fund",
			},
			expected: Goals{
				Name: "Emergency Fund",
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			assert.Equal(suite.T(), tt.expected, tt.goal)
		})
	}
}

func (suite *ProfileTestSuite) TestProfileGoals() {
	goals2Years := []Goals{{Name: "Save 10k", TargetAmount: 10000}}
	goals5Years := []Goals{{Name: "Buy car", TargetAmount: 25000}}
	goals10Years := []Goals{{Name: "Retirement", TargetAmount: 500000}}

	profileGoals := ProfileGoals{
		Goals2Years:  goals2Years,
		Goals5Years:  goals5Years,
		Goals10Years: goals10Years,
	}

	assert.Equal(suite.T(), goals2Years, profileGoals.Goals2Years)
	assert.Equal(suite.T(), goals5Years, profileGoals.Goals5Years)
	assert.Equal(suite.T(), goals10Years, profileGoals.Goals10Years)
}

func (suite *ProfileTestSuite) TestProfile() {
	now := time.Now()
	tests := []struct {
		name     string
		profile  Profile
		isValid  bool
	}{
		{
			name: "Complete Profile",
			profile: Profile{
				ID:             "123",
				FirstName:      "John",
				LastName:       "Doe",
				Email:          "john@example.com",
				Phone:          "123456789",
				BirthDate:      "1990-01-01",
				Sexo:           "M",
				Cep:            "12345-678",
				City:           "São Paulo",
				State:          "SP",
				UserProviderID: "provider123",
				Profession: ProfileProfession{
					Profession:   "Developer",
					ContractType: ContractTypeCTL,
				},
				Goals: ProfileGoals{
					Goals2Years: []Goals{{Name: "Emergency Fund"}},
				},
				CreatedAt: now,
				UpdatedAt: now,
			},
			isValid: true,
		},
		{
			name: "Minimal Profile",
			profile: Profile{
				FirstName: "Jane",
				LastName:  "Smith",
				Email:     "jane@example.com",
				Sexo:      "F",
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.isValid {
				assert.NotEmpty(suite.T(), tt.profile.FirstName)
				assert.NotEmpty(suite.T(), tt.profile.LastName)
				assert.NotEmpty(suite.T(), tt.profile.Email)
			}
		})
	}
}

func (suite *ProfileTestSuite) TestValidateBirthDate() {
	profile := &Profile{}

	tests := []struct {
		name        string
		birthDate   string
		expectError bool
		expected    string
	}{
		{
			name:        "Valid date",
			birthDate:   "1990-05-15",
			expectError: false,
			expected:    "1990-05-15",
		},
		{
			name:        "Empty date",
			birthDate:   "",
			expectError: true,
		},
		{
			name:        "Invalid format",
			birthDate:   "15/05/1990",
			expectError: true,
		},
		{
			name:        "Invalid date",
			birthDate:   "2000-13-32",
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result, err := profile.ValidateBirthDate(tt.birthDate)
			
			if tt.expectError {
				assert.Error(suite.T(), err)
			} else {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), tt.expected, result.Format("2006-01-02"))
			}
		})
	}
}

func (suite *ProfileTestSuite) TestSetBirthDate() {
	tests := []struct {
		name        string
		birthDate   string
		expectError bool
		expected    string
	}{
		{
			name:        "Valid date",
			birthDate:   "1990-05-15",
			expectError: false,
			expected:    "1990-05-15",
		},
		{
			name:        "Empty date",
			birthDate:   "",
			expectError: true,
		},
		{
			name:        "Invalid format",
			birthDate:   "15/05/1990",
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			profile := &Profile{}
			err := profile.SetBirthDate(tt.birthDate)
			
			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Empty(suite.T(), profile.BirthDate)
			} else {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), tt.expected, profile.BirthDate)
			}
		})
	}
}

func (suite *ProfileTestSuite) TestValidate() {
	profile := &Profile{}
	err := profile.Validate()
	assert.NoError(suite.T(), err)
}

// Benchmark tests for performance
func BenchmarkProfileCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		profile := Profile{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Sexo:      "M",
		}
		_ = profile
	}
}

func BenchmarkValidateBirthDate(b *testing.B) {
	profile := &Profile{}
	birthDate := "1990-05-15"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = profile.ValidateBirthDate(birthDate)
	}
}

func BenchmarkSetBirthDate(b *testing.B) {
	birthDate := "1990-05-15"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profile := &Profile{}
		_ = profile.SetBirthDate(birthDate)
	}
}