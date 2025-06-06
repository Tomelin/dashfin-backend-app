package repository

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	entity_profile "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ProfileProfessionRepositoryTestSuite struct {
	suite.Suite
	mockDB *MockFirebaseDB
	repo   ProfileProfessionRepositoryInterface
	ctx    context.Context
}

func TestProfileProfessionRepositorySuite(t *testing.T) {
	suite.Run(t, new(ProfileProfessionRepositoryTestSuite))
}

func (suite *ProfileProfessionRepositoryTestSuite) SetupTest() {
	suite.mockDB = new(MockFirebaseDB)
	suite.ctx = context.Background()
	
	repo, err := InicializeProfileProfessionRepository(suite.mockDB)
	assert.NoError(suite.T(), err)
	suite.repo = repo
}

func (suite *ProfileProfessionRepositoryTestSuite) TestInicializeProfileProfessionRepository() {
	tests := []struct {
		name        string
		db          *MockFirebaseDB
		expectError bool
	}{
		{
			name:        "Valid DB",
			db:          suite.mockDB,
			expectError: false,
		},
		{
			name:        "Nil DB",
			db:          nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			repo, err := InicializeProfileProfessionRepository(tt.db)
			
			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), repo)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), repo)
			}
		})
	}
}

func (suite *ProfileProfessionRepositoryTestSuite) TestUpdateProfileProfession() {
	profile := &entity_profile.Profile{
		ID:             "123",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		UserProviderID: "provider123",
		Profession: entity_profile.ProfileProfession{
			Profession:    "Software Engineer",
			Company:       "TechCorp",
			ContractType:  entity_profile.ContractTypeCTL,
			MonthlyIncome: 5000.0,
		},
	}

	profiles := []entity_profile.Profile{*profile}

	tests := []struct {
		name         string
		profile      *entity_profile.Profile
		mockError    error
		filterReturn []byte
		filterError  error
		expectError  bool
	}{
		{
			name:         "Successful update",
			profile:      profile,
			mockError:    nil,
			filterReturn: mustMarshal(profiles),
			filterError:  nil,
			expectError:  false,
		},
		{
			name:        "Nil profile",
			profile:     nil,
			expectError: true,
		},
		{
			name:        "Update error",
			profile:     profile,
			mockError:   errors.New("update failed"),
			expectError: true,
		},
		{
			name:         "Filter error after update",
			profile:      profile,
			mockError:    nil,
			filterError:  errors.New("filter failed"),
			expectError:  true,
		},
		{
			name:         "No user found after update",
			profile:      profile,
			mockError:    nil,
			filterReturn: []byte(`[]`),
			filterError:  nil,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.profile != nil {
				suite.mockDB.On("Update", suite.ctx, tt.profile.ID, mock.Anything, "profiles").
					Return(tt.mockError).Once()
				
				if tt.mockError == nil {
					expectedFilter := map[string]interface{}{
						"userProviderID": tt.profile.UserProviderID,
					}
					suite.mockDB.On("GetByFilter", suite.ctx, expectedFilter, "profiles").
						Return(tt.filterReturn, tt.filterError).Once()
				}
			}

			result, err := suite.repo.UpdateProfileProfession(suite.ctx, tt.profile)
			
			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), result)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), result)
				assert.Equal(suite.T(), tt.profile.FirstName, result.FirstName)
				assert.Equal(suite.T(), "Software Engineer", result.Profession.Profession)
				assert.Equal(suite.T(), "TechCorp", result.Profession.Company)
				assert.Equal(suite.T(), entity_profile.ContractTypeCTL, result.Profession.ContractType)
				assert.Equal(suite.T(), 5000.0, result.Profession.MonthlyIncome)
			}

			suite.mockDB.AssertExpectations(suite.T())
		})
	}
}

func (suite *ProfileProfessionRepositoryTestSuite) TestGetByFilter() {
	profiles := []entity_profile.Profile{
		{
			ID:        "123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Profession: entity_profile.ProfileProfession{
				Profession:    "Software Engineer",
				Company:       "TechCorp",
				ContractType:  entity_profile.ContractTypeCTL,
				MonthlyIncome: 5000.0,
			},
		},
	}

	tests := []struct {
		name        string
		filter      map[string]interface{}
		mockReturn  []byte
		mockError   error
		expectError bool
	}{
		{
			name:        "Successful filter",
			filter:      map[string]interface{}{"email": "john@example.com"},
			mockReturn:  mustMarshal(profiles),
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "Filter by profession",
			filter:      map[string]interface{}{"profession.profession": "Software Engineer"},
			mockReturn:  mustMarshal(profiles),
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "Filter by contract type",
			filter:      map[string]interface{}{"profession.contractType": "clt"},
			mockReturn:  mustMarshal(profiles),
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "Nil filter",
			filter:      nil,
			expectError: true,
		},
		{
			name:        "Database error",
			filter:      map[string]interface{}{"email": "john@example.com"},
			mockError:   errors.New("database error"),
			expectError: true,
		},
		{
			name:        "Invalid JSON response",
			filter:      map[string]interface{}{"email": "john@example.com"},
			mockReturn:  []byte(`invalid json`),
			mockError:   nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.filter != nil && tt.mockReturn != nil {
				suite.mockDB.On("GetByFilter", suite.ctx, tt.filter, "profiles").
					Return(tt.mockReturn, tt.mockError).Once()
			}

			result, err := suite.repo.GetByFilter(suite.ctx, tt.filter)
			
			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), result)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), result)
				assert.Len(suite.T(), result, 1)
				assert.Equal(suite.T(), "John", result[0].FirstName)
				assert.Equal(suite.T(), "Software Engineer", result[0].Profession.Profession)
			}

			suite.mockDB.AssertExpectations(suite.T())
		})
	}
}

// Test different contract types
func (suite *ProfileProfessionRepositoryTestSuite) TestContractTypes() {
	contractTypes := []struct {
		name         string
		contractType entity_profile.ContractTypeValues
	}{
		{"CTL", entity_profile.ContractTypeCTL},
		{"PJ", entity_profile.ContractTypePJ},
		{"Temporary", entity_profile.ContractTypeTemporary},
		{"Internship", entity_profile.ContractTypeInternship},
		{"Public Servant", entity_profile.ContractTypePublicServant},
		{"Other", entity_profile.ContractTypeOther},
	}

	for _, tt := range contractTypes {
		suite.Run(tt.name, func() {
			profile := &entity_profile.Profile{
				ID:             "123",
				UserProviderID: "provider123",
				Profession: entity_profile.ProfileProfession{
					Profession:   "Test Profession",
					ContractType: tt.contractType,
				},
			}

			profiles := []entity_profile.Profile{*profile}

			suite.mockDB.On("Update", suite.ctx, profile.ID, mock.Anything, "profiles").
				Return(nil).Once()
			
			expectedFilter := map[string]interface{}{
				"userProviderID": profile.UserProviderID,
			}
			suite.mockDB.On("GetByFilter", suite.ctx, expectedFilter, "profiles").
				Return(mustMarshal(profiles), nil).Once()

			result, err := suite.repo.UpdateProfileProfession(suite.ctx, profile)
			
			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), result)
			assert.Equal(suite.T(), tt.contractType, result.Profession.ContractType)

			suite.mockDB.AssertExpectations(suite.T())
		})
	}
}

// Benchmark tests
func BenchmarkUpdateProfileProfession(b *testing.B) {
	mockDB := new(MockFirebaseDB)
	repo, _ := InicializeProfileProfessionRepository(mockDB)
	ctx := context.Background()
	
	profile := &entity_profile.Profile{
		ID:             "123",
		UserProviderID: "provider123",
		Profession: entity_profile.ProfileProfession{
			Profession:    "Software Engineer",
			Company:       "TechCorp",
			ContractType:  entity_profile.ContractTypeCTL,
			MonthlyIncome: 5000.0,
		},
	}

	profiles := []entity_profile.Profile{*profile}

	mockDB.On("Update", ctx, mock.Anything, mock.Anything, "profiles").Return(nil)
	mockDB.On("GetByFilter", ctx, mock.Anything, "profiles").
		Return(mustMarshal(profiles), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.UpdateProfileProfession(ctx, profile)
	}
}

func BenchmarkProfessionGetByFilter(b *testing.B) {
	mockDB := new(MockFirebaseDB)
	repo, _ := InicializeProfileProfessionRepository(mockDB)
	ctx := context.Background()
	
	profiles := []entity_profile.Profile{
		{
			ID: "123", 
			Profession: entity_profile.ProfileProfession{
				Profession:   "Software Engineer",
				ContractType: entity_profile.ContractTypeCTL,
			},
		},
	}
	
	filter := map[string]interface{}{"profession.profession": "Software Engineer"}

	mockDB.On("GetByFilter", ctx, filter, "profiles").
		Return(mustMarshal(profiles), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByFilter(ctx, filter)
	}
}