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

type ProfileGoalsRepositoryTestSuite struct {
	suite.Suite
	mockDB *MockFirebaseDB
	repo   ProfileGoalsRepositoryInterface
	ctx    context.Context
}

func TestProfileGoalsRepositorySuite(t *testing.T) {
	suite.Run(t, new(ProfileGoalsRepositoryTestSuite))
}

func (suite *ProfileGoalsRepositoryTestSuite) SetupTest() {
	suite.mockDB = new(MockFirebaseDB)
	suite.ctx = context.Background()
	
	repo, err := InicializeProfileGoalsRepository(suite.mockDB)
	assert.NoError(suite.T(), err)
	suite.repo = repo
}

func (suite *ProfileGoalsRepositoryTestSuite) TestInicializeProfileGoalsRepository() {
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
			repo, err := InicializeProfileGoalsRepository(tt.db)
			
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

func (suite *ProfileGoalsRepositoryTestSuite) TestUpdateProfileGoals() {
	profile := &entity_profile.Profile{
		ID:             "123",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		UserProviderID: "provider123",
		Goals: entity_profile.ProfileGoals{
			Goals2Years: []entity_profile.Goals{
				{Name: "Emergency Fund", TargetAmount: 10000},
			},
			Goals5Years: []entity_profile.Goals{
				{Name: "Buy Car", TargetAmount: 25000},
			},
			Goals10Years: []entity_profile.Goals{
				{Name: "Retirement", TargetAmount: 500000},
			},
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

			result, err := suite.repo.UpdateProfileGoals(suite.ctx, tt.profile)
			
			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), result)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), result)
				assert.Equal(suite.T(), tt.profile.FirstName, result.FirstName)
				assert.Len(suite.T(), result.Goals.Goals2Years, 1)
				assert.Equal(suite.T(), "Emergency Fund", result.Goals.Goals2Years[0].Name)
			}

			suite.mockDB.AssertExpectations(suite.T())
		})
	}
}

func (suite *ProfileGoalsRepositoryTestSuite) TestGetByFilter() {
	profiles := []entity_profile.Profile{
		{
			ID:        "123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Goals: entity_profile.ProfileGoals{
				Goals2Years: []entity_profile.Goals{
					{Name: "Emergency Fund", TargetAmount: 10000},
				},
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
				assert.Len(suite.T(), result[0].Goals.Goals2Years, 1)
			}

			suite.mockDB.AssertExpectations(suite.T())
		})
	}
}

// Benchmark tests
func BenchmarkUpdateProfileGoals(b *testing.B) {
	mockDB := new(MockFirebaseDB)
	repo, _ := InicializeProfileGoalsRepository(mockDB)
	ctx := context.Background()
	
	profile := &entity_profile.Profile{
		ID:             "123",
		UserProviderID: "provider123",
		Goals: entity_profile.ProfileGoals{
			Goals2Years: []entity_profile.Goals{
				{Name: "Emergency Fund", TargetAmount: 10000},
			},
		},
	}

	profiles := []entity_profile.Profile{*profile}

	mockDB.On("Update", ctx, mock.Anything, mock.Anything, "profiles").Return(nil)
	mockDB.On("GetByFilter", ctx, mock.Anything, "profiles").
		Return(mustMarshal(profiles), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.UpdateProfileGoals(ctx, profile)
	}
}

func BenchmarkGoalsGetByFilter(b *testing.B) {
	mockDB := new(MockFirebaseDB)
	repo, _ := InicializeProfileGoalsRepository(mockDB)
	ctx := context.Background()
	
	profiles := []entity_profile.Profile{
		{
			ID: "123", 
			Goals: entity_profile.ProfileGoals{
				Goals2Years: []entity_profile.Goals{{Name: "Test"}},
			},
		},
	}
	
	filter := map[string]interface{}{"email": "john@example.com"}

	mockDB.On("GetByFilter", ctx, filter, "profiles").
		Return(mustMarshal(profiles), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByFilter(ctx, filter)
	}
}